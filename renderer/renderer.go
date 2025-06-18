package renderer

import (
	. "ThreeDView/camera"
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

type threeDWidgetInterface interface {
	ThreeDWidgetInterface
	GetCamera() *Camera
	GetObjects() []*Object
}

type Renderer struct {
	widget        threeDWidgetInterface
	img           *image.RGBA
	zBuffer       [][]float64
	renderWorkers []*renderWorker
	workerChannel chan *instruction
}

func NewRenderer(widget threeDWidgetInterface) *Renderer {
	runtime.GOMAXPROCS(runtime.NumCPU())
	renderer := &Renderer{
		widget:        widget,
		img:           nil,
		zBuffer:       nil,
		renderWorkers: make([]*renderWorker, 2),
		workerChannel: make(chan *instruction),
	}

	for i := range renderer.renderWorkers {
		renderer.renderWorkers[i] = newRenderWorker(renderer.widget, renderer.workerChannel)
		go renderer.renderWorkers[i].start()
	}

	go func() {
		for {
			var states []string
			for _, worker := range renderer.renderWorkers {
				states = append(states, worker.State)
			}
			log.Println("Render worker states:", states)
			var successCounts []int
			for _, worker := range renderer.renderWorkers {
				successCounts = append(successCounts, worker.Success)
			}
			log.Println("Render worker success counts:", successCounts)

			time.Sleep(2 * time.Second)

		}
	}()

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

func (r *Renderer) getFacesInFrustum() chan interface{} {
	callbackChannel := make(chan interface{})
	wg := sync.WaitGroup{}
	wg.Add(len(r.widget.GetObjects()))
	go func() {
		for _, object := range r.widget.GetObjects() {
			r.workerChannel <- &instruction{instructionType: "getFacesInFrustum", data: object.GetFaces(), callbackChannel: callbackChannel, doneFunction: func() { wg.Done() }}
		}
	}()
	go func() {
		wg.Wait()
		close(callbackChannel)
	}()
	return callbackChannel
}

func (r *Renderer) clipAndProjectFaces() []ProjectedFaceData {
	callbackChannel := make(chan interface{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for faceData := range r.getFacesInFrustum() {
			face, _ := faceData.(FaceData)
			wg.Add(1)
			r.workerChannel <- &instruction{instructionType: "clipAndProject", data: face, callbackChannel: callbackChannel, doneFunction: func() {
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
			img.Set(x, y, color.Gray{Y: gray})
		}
	}
	r.img = img
}

func (r *Renderer) Render() image.Image {
	r.setupImg()
	if len(r.widget.GetObjects()) == 0 {
		return r.img
	}
	r.resetZBuffer()
	faces := r.clipAndProjectFaces()
	r.renderColors(faces)
	r.renderFaceOutlines(faces)
	r.renderZBuffer()
	return r.img
}
