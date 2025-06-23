package renderer

import (
	. "ThreeDView/object"
	. "ThreeDView/types"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Renderer struct {
	widget        ThreeDWidgetInterface
	img           *image.RGBA
	zBuffer       [][]float64
	renderWorkers []*renderWorker
	workerChannel chan *instruction
}

func NewRenderer(widget ThreeDWidgetInterface) *Renderer {
	runtime.GOMAXPROCS(runtime.NumCPU())
	renderer := &Renderer{
		widget:        widget,
		img:           nil,
		zBuffer:       nil,
		renderWorkers: make([]*renderWorker, runtime.NumCPU()),
		workerChannel: make(chan *instruction, 1000),
	}

	for i := range renderer.renderWorkers {
		renderer.renderWorkers[i] = newRenderWorker(renderer.widget, renderer.workerChannel)
		go renderer.renderWorkers[i].start()
	}
	return renderer
}

func (r *Renderer) setupImg() {
	r.img = image.NewRGBA(image.Rect(0, 0, int(r.widget.GetWidth()), int(r.widget.GetHeight())))
	draw.Draw(r.img, r.img.Bounds(), &image.Uniform{C: r.widget.GetBackgroundColor()}, image.Point{}, draw.Src)
}

func (r *Renderer) resetZBuffer() {
	width := int(r.widget.GetWidth())
	height := int(r.widget.GetHeight())
	r.zBuffer = make([][]float64, width)
	for i := range r.zBuffer {
		r.zBuffer[i] = make([]float64, height)
		for j := range r.zBuffer[i] {
			r.zBuffer[i][j] = math.Inf(1)
		}
	}
}

func (r *Renderer) clipAndProjectFaces() []ProjectedFaceData {
	callbackChannel := make(chan interface{}, 10000)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for faceData := range r.widget.GetCamera().GetVisibleFaces() {
			wg.Add(1)
			r.workerChannel <- &instruction{instructionType: "clipAndProject", data: faceData, callbackChannel: callbackChannel, doneFunction: func() {
				wg.Done()
			}}
		}
		wg.Done()
	}()
	go func() {
		wg.Wait()
		close(callbackChannel)
	}()
	var projectedFaces []ProjectedFaceData
	for faceData := range callbackChannel {
		face, _ := faceData.(ProjectedFaceData)
		projectedFaces = append(projectedFaces, face)
	}
	return projectedFaces
}

func (r *Renderer) renderColors(faces []ProjectedFaceData) {
	if r.widget.GetRenderFaceColors() {
		for _, face := range faces {
			drawFilledTriangle(r.img, face, r.zBuffer, face.HasTexture && r.widget.GetRenderTextures())
		}
	}
}

func (r *Renderer) renderFaceOutlines(faces []ProjectedFaceData) {
	if r.widget.GetRenderFaceOutlines() {
		for _, face := range faces {
			var outlineColor color.Color
			if !r.widget.GetRenderFaceColors() {
				outlineColor = face.Color
			} else {
				outlineColor = color.Black
			}
			drawEdge(r.img, face.Face[0], face.Z[0], face.Face[1], face.Z[1], outlineColor, r.zBuffer)
			drawEdge(r.img, face.Face[1], face.Z[1], face.Face[2], face.Z[2], outlineColor, r.zBuffer)
			drawEdge(r.img, face.Face[2], face.Z[2], face.Face[0], face.Z[0], outlineColor, r.zBuffer)
		}
	}
}

func (r *Renderer) renderZBuffer() {
	if !r.widget.GetRenderZBuffer() {
		return
	}
	img := image.NewRGBA(r.img.Bounds())
	var logZVals []float64
	for x := 0; x < r.img.Bounds().Dx(); x++ {
		for y := 0; y < r.img.Bounds().Dy(); y++ {
			if len(r.zBuffer) <= x || len(r.zBuffer[x]) <= y {
				continue
			}
			z := r.zBuffer[x][y]
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				logZVals = append(logZVals, math.Log(z))
			}
		}
	}
	if len(logZVals) == 0 {
		r.img = img
	}
	sort.Float64s(logZVals)
	lowIdx := int(0.02 * float64(len(logZVals)))
	highIdx := int(0.98 * float64(len(logZVals)))
	if highIdx <= lowIdx {
		lowIdx = 0
		highIdx = len(logZVals) - 1
	}
	if len(logZVals) == 0 {
		r.img = img
	}
	minZ, maxZ := logZVals[lowIdx], logZVals[highIdx]
	if minZ == maxZ {
		minZ, maxZ = 0, 1
	}
	for x := 0; x < r.img.Bounds().Dx(); x++ {
		for y := 0; y < r.img.Bounds().Dy(); y++ {
			if len(r.zBuffer) <= x || len(r.zBuffer[x]) <= y {
				log.Println("Skipping pixel due to out of bounds zBuffer access at", x, y)
				continue
			}
			z := r.zBuffer[x][y]
			var gray uint8 = 255
			if !math.IsInf(z, 1) && z > 0 {
				logZ := math.Log(z)
				norm := (logZ - minZ) / (maxZ - minZ)
				if norm < 0 {
					norm = 0
				}
				if norm > 1 {
					norm = 1
				}
				gray = uint8(norm * 255)
			}
			img.Set(x, y, color.Gray{Y: 255 - gray})
		}
	}
	r.img = img
}

func (r *Renderer) renderPseudoShading() {
	if !r.widget.GetRenderPseudoShading() {
		return
	}

	const minShadeFactor = 0.7
	const maxShadeFactor = 1.0

	var zVals []float64
	for x := 0; x < r.img.Bounds().Dx(); x++ {
		for y := 0; y < r.img.Bounds().Dy(); y++ {
			if len(r.zBuffer) <= x || len(r.zBuffer[x]) <= y {
				continue
			}
			z := r.zBuffer[x][y]
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				zVals = append(zVals, z)
			}
		}
	}
	if len(zVals) == 0 {
		return
	}
	sort.Float64s(zVals)
	lowIdx := int(0.02 * float64(len(zVals)))
	highIdx := int(0.98 * float64(len(zVals)))
	if highIdx <= lowIdx {
		lowIdx = 0
		highIdx = len(zVals) - 1
	}
	if len(zVals) == 0 {
		return
	}
	minZ, maxZ := zVals[lowIdx], zVals[highIdx]
	if minZ == maxZ {
		minZ, maxZ = 0, 1
	}

	for x := 0; x < r.img.Bounds().Dx(); x++ {
		for y := 0; y < r.img.Bounds().Dy(); y++ {
			if len(r.zBuffer) <= x || len(r.zBuffer[x]) <= y {
				log.Println("Skipping pixel due to out of bounds zBuffer access at", x, y)
				continue
			}
			z := r.zBuffer[x][y]

			if !math.IsInf(z, 1) && z > 0 {
				norm := (z - minZ) / (maxZ - minZ)
				if norm < 0 {
					norm = 0
				}
				if norm > 1 {
					norm = 1
				}

				shadeFactor := maxShadeFactor - (norm * (maxShadeFactor - minShadeFactor))

				originalColor := r.img.At(x, y).(color.RGBA)
				newColor := color.RGBA{
					R: uint8(float64(originalColor.R) * shadeFactor),
					G: uint8(float64(originalColor.G) * shadeFactor),
					B: uint8(float64(originalColor.B) * shadeFactor),
					A: originalColor.A,
				}
				r.img.Set(x, y, newColor)
			}
		}
	}
}

func (r *Renderer) renderEdgeOutlines() {
	if !r.widget.GetRenderEdgeOutlines() {
		return
	}
	edgeMask := detectZBufferEdges(r.zBuffer, 0.05, 0.1, 5.0, 0.5)
	thickness := 1
	outlineColor := color.Black
	for x := 0; x < r.img.Bounds().Dx(); x++ {
		for y := 0; y < r.img.Bounds().Dy(); y++ {
			if edgeMask[x][y] {
				for dx := -thickness; dx <= thickness; dx++ {
					for dy := -thickness; dy <= thickness; dy++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < r.img.Bounds().Dx() && ny >= 0 && ny < r.img.Bounds().Dy() {
							r.img.Set(nx, ny, outlineColor)
						}
					}
				}
			}
		}
	}
}

func (r *Renderer) Render() image.Image {
	r.setupImg()
	if len(r.widget.GetObjects()) == 0 {
		return r.img
	}
	r.resetZBuffer()
	startTime1 := time.Now()
	faces := r.clipAndProjectFaces()
	log.Println("Projection and clipping took", time.Since(startTime1))
	startTime2 := time.Now()
	r.renderColors(faces)
	r.renderFaceOutlines(faces)
	r.renderZBuffer()
	r.renderEdgeOutlines()
	r.renderPseudoShading()
	log.Println("Rendering took", time.Since(startTime2))
	log.Println("FPS:", int(1.0/time.Since(startTime1).Seconds()))
	return r.img
}
