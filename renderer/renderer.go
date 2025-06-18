package renderer

import (
	. "ThreeDView/camera"
	. "ThreeDView/object"
	. "ThreeDView/types"
	"image"
	"image/draw"
	"log"
	"math"
	"runtime"
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
		renderWorkers: make([]*renderWorker, runtime.NumCPU()),
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
	r.zBuffer = make([][]float64, height)
	for i := range r.zBuffer {
		r.zBuffer[i] = make([]float64, width)
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

func (r *Renderer) clipAndProjectFaces() chan interface{} {
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
	return callbackChannel
}

func (r *Renderer) Render() image.Image {
	r.setupImg()
	r.resetZBuffer()
	if r.widget.GetRenderFaceColors() {
		for faceData := range r.clipAndProjectFaces() {
			face, _ := faceData.(ProjectedFaceData)
			drawFilledTriangle(r.img, face, r.zBuffer, face.HasTexture && r.widget.GetRenderTextures())
		}
	}
	return r.img
}
