package ThreeDView

import (
	. "ThreeDView/camera"
	. "ThreeDView/object"
	. "ThreeDView/types"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

var (
	Width            = Pixel(800)
	Height           = Pixel(600)
	resolutionFactor = 1.0
)

// ThreeDWidget is a widget that displays 3D objects
type ThreeDWidget struct {
	widget.BaseWidget
	image              *canvas.Image // The image that is rendered on
	camera             *Camera       // The camera of the 3D widget
	objects            []*Object     // The objects in the 3D widget
	tickMethods        []func()      // The methods that are called every frame
	bgColor            color.Color   // The background color of the 3D widget
	renderFaceOutlines bool          // Whether the faces should be rendered with outlines
	renderFaceColors   bool          // Whether the faces should be rendered with colors
	fpsCap             float64       // The maximum frames per second the widget should render at
	tpsCap             float64       // The maximum ticks per second the widget should tick at
}

// NewThreeDWidget creates a new 3D widget
func NewThreeDWidget() *ThreeDWidget {
	w := &ThreeDWidget{}
	w.ExtendBaseWidget(w)
	w.bgColor = color.Transparent
	w.renderFaceOutlines = false
	w.renderFaceColors = true
	standardCamera := NewCamera(mgl.Vec3{}, mgl.QuatIdent())
	w.camera = &standardCamera
	w.objects = []*Object{}
	w.image = canvas.NewImageFromImage(w.render())
	w.fpsCap = math.Inf(1)
	w.tpsCap = math.Inf(1)
	go w.renderLoop()
	go w.tickLoop()
	return w
}

func (w *ThreeDWidget) tickLoop() {
	for {
		if w.tpsCap == 0 || !w.Visible() {
			continue
		}
		startTime := time.Now()
		tickDuration := time.Second / time.Duration(w.tpsCap)

		for _, tickMethod := range w.tickMethods {
			tickMethod()
		}

		elapsedTime := time.Since(startTime)
		if elapsedTime < tickDuration {
			time.Sleep(tickDuration - elapsedTime)
		}
		if elapsedTime > tickDuration && tickDuration != 0 {
			log.Println("WARNING: Tick took too long to execute (", elapsedTime, " > ", tickDuration, ")")
		}
	}
}

func (w *ThreeDWidget) renderLoop() {
	for {
		if w.fpsCap == 0 || !w.Visible() {
			continue
		}
		frameStartTime := time.Now()
		frameDuration := time.Second / time.Duration(w.fpsCap)

		w.image.Image = w.render()
		fyne.Do(func() {
			canvas.Refresh(w.image)
		})

		frameElapsedTime := time.Since(frameStartTime)
		if frameElapsedTime < frameDuration {
			time.Sleep(frameDuration - frameElapsedTime)
		}
		if frameElapsedTime > frameDuration && frameDuration != 0 {
			log.Println("WARNING: Frame took too long to render (", frameElapsedTime, " > ", frameDuration, ")")
		}
	}
}

// RegisterTickMethod registers an animation function to be called every frame
func (w *ThreeDWidget) RegisterTickMethod(tick func()) {
	w.tickMethods = append(w.tickMethods, tick)
}

// AddObject adds a 3D object as Object to the widget. This should be called in the method that creates the object
func (w *ThreeDWidget) AddObject(object *Object) {
	w.objects = append(w.objects, object)
}

func (w *ThreeDWidget) GetCamera() *Camera {
	return w.camera
}

func (w *ThreeDWidget) GetWidth() Pixel {
	return Width
}

func (w *ThreeDWidget) GetHeight() Pixel {
	return Height
}

// SetCamera sets the camera of the 3D widget
func (w *ThreeDWidget) SetCamera(camera *Camera) {
	w.camera = camera
}

// SetBackgroundColor sets the background color of the 3D widget
func (w *ThreeDWidget) SetBackgroundColor(color color.Color) {
	w.bgColor = color
}

// SetFPSCap sets the maximum frames per second the widget should render at
func (w *ThreeDWidget) SetFPSCap(fps float64) {
	w.fpsCap = fps
}

// SetTPSCap sets the maximum ticks per second the widget should update at. Animations are triggered at this rate
func (w *ThreeDWidget) SetTPSCap(tps float64) {
	w.tpsCap = tps
}

// SetResolutionFactor sets the resolution factor of the 3D widget. This is a factor that is multiplied with the size of the widget to determine the resolution of the 3D rendering
func (w *ThreeDWidget) SetResolutionFactor(factor float64) {
	resolutionFactor = factor
}

// SetRenderFaceOutlines sets whether the faces should be rendered with outlines.
// If false, only colors will be rendered. If colors are also false, nothing will be rendered.
// If true, the faces will be rendered with black outlines or the color of the face if face colors are disabled.
// Default is false
func (w *ThreeDWidget) SetRenderFaceOutlines(newVal bool) {
	w.renderFaceOutlines = newVal
}

// SetRenderFaceColors sets whether the faces should be rendered with colors.
// If false, only outlines will be rendered. If outline is also false, nothing will be rendered.
// Default is true
func (w *ThreeDWidget) SetRenderFaceColors(newVal bool) {
	w.renderFaceColors = newVal
}

func (w *ThreeDWidget) CreateRenderer() fyne.WidgetRenderer {
	return &threeDRenderer{image: w.image}
}

func (w *ThreeDWidget) render() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(Width), int(Height)))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: w.bgColor}, image.Point{}, draw.Src)

	var faces []FaceData
	var wg3d sync.WaitGroup
	var mu3d sync.Mutex
	wg3d.Add(len(w.objects))
	for _, object := range w.objects {
		go func(object *Object) {
			defer wg3d.Done()
			objectFaces := object.GetFaces()
			mu3d.Lock()
			for _, face := range objectFaces {
				if w.camera.FaceOverlapsFrustum(face.Face, Width, Height) {
					faces = append(faces, face)
				}
			}
			mu3d.Unlock()
		}(object)
	}
	wg3d.Wait()

	var projectedFaces []ProjectedFaceData
	var wg2d sync.WaitGroup
	wg2d.Add(len(faces))
	var mu sync.Mutex
	for _, face := range faces {
		go func(face FaceData) {
			defer wg2d.Done()
			p1 := w.camera.Project(face.Face[0], Width, Height)
			p2 := w.camera.Project(face.Face[1], Width, Height)
			p3 := w.camera.Project(face.Face[2], Width, Height)

			if !triangleOverlapsScreen(p1, p2, p3, Width, Height) {
				return
			}

			mu.Lock()
			projectedFaces = append(projectedFaces, ProjectedFaceData{Face: [3]mgl.Vec2{p1, p2, p3}, Color: face.Color, Distance: face.Distance})
			mu.Unlock()
		}(face)
	}
	wg2d.Wait()

	sort.Slice(projectedFaces, func(i, j int) bool {
		return projectedFaces[i].Distance > projectedFaces[j].Distance
	})

	for _, face := range projectedFaces {
		drawFace(img, face, w.renderFaceOutlines, w.renderFaceColors)
	}

	return img
}

func (w *ThreeDWidget) Dragged(event *fyne.DragEvent) {
	if controller, ok := w.camera.Controller.(DragController); ok {
		controller.OnDrag(event.Dragged.DX, event.Dragged.DY)
	}
}

func (w *ThreeDWidget) DragEnd() {
	if controller, ok := w.camera.Controller.(DragController); ok {
		controller.OnDragEnd()
	}
}

func (w *ThreeDWidget) Scrolled(event *fyne.ScrollEvent) {
	if controller, ok := w.camera.Controller.(ScrollController); ok {
		controller.OnScroll(event.Scrolled.DX, event.Scrolled.DY)
	}
}

type threeDRenderer struct {
	image *canvas.Image
}

// Layout resizes the widget to the given size
func (r *threeDRenderer) Layout(size fyne.Size) {
	r.image.Resize(size)
	Width = Pixel(float64(size.Width) * resolutionFactor)
	Height = Pixel(float64(size.Height) * resolutionFactor)
}

// MinSize returns the minimum size of the widget
func (r *threeDRenderer) MinSize() fyne.Size {
	return r.image.MinSize()
}

// Refresh refreshes the widget
func (r *threeDRenderer) Refresh() {
	canvas.Refresh(r.image)
}

// Objects returns the objects of the widget. This will be only the image that is rendered on
func (r *threeDRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.image}
}

func (r *threeDRenderer) Destroy() {}

func drawFace(img *image.RGBA, face ProjectedFaceData, renderFaceOutlines bool, renderFaceColors bool) {
	if renderFaceColors {
		drawFilledTriangle(img, face.Face[0], face.Face[1], face.Face[2], face.Color)
	}

	if !renderFaceOutlines {
		log.Println("drawFace: skipping outlines")
		return
	}
	var outlineColor color.Color
	if !renderFaceColors {
		outlineColor = face.Color
	} else {
		outlineColor = color.Black
	}
	point1 := face.Face[0]
	point2 := face.Face[1]
	point3 := face.Face[2]
	drawLine(img, point1, point2, outlineColor)
	drawLine(img, point2, point3, outlineColor)
	drawLine(img, point3, point1, outlineColor)
}

func drawLine(img *image.RGBA, point1, point2 mgl.Vec2, lineColor color.Color) {
	x0 := int(math.Round(point1.X()))
	y0 := int(math.Round(point1.Y()))
	x1 := int(math.Round(point2.X()))
	y1 := int(math.Round(point2.Y()))
	if math.IsNaN(point1.X()) || math.IsNaN(point1.Y()) || math.IsNaN(point2.X()) || math.IsNaN(point2.Y()) ||
		math.IsInf(point1.X(), 0) || math.IsInf(point1.Y(), 0) || math.IsInf(point2.X(), 0) || math.IsInf(point2.Y(), 0) {
		log.Println("drawLine: NaN or Inf detected, skipping line")
		return
	}
	dx := int(math.Abs(float64(x1 - x0)))
	dy := int(math.Abs(float64(y1 - y0)))
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy
	maxIter := dx + dy + 10
	iter := 0
	for {
		if x0 >= 0 && x0 < int(Width) && y0 >= 0 && y0 < int(Height) {
			img.Set(x0, y0, lineColor)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
		iter++
		if iter > maxIter {
			log.Println("drawLine: max iterations reached, breaking to avoid infinite loop")
			break
		}
	}
}

func drawFilledTriangle(img *image.RGBA, p1, p2, p3 mgl.Vec2, fillColor color.Color) {
	if p2.Y() < p1.Y() {
		p1, p2 = p2, p1
	}
	if p3.Y() < p1.Y() {
		p1, p3 = p3, p1
	}
	if p3.Y() < p2.Y() {
		p2, p3 = p3, p2
	}
	drawHorizontalLine := func(y, x1, x2 Pixel, color color.Color) {
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		for x := x1; x <= x2; x++ {
			if y >= 0 && y < Height && x >= 0 && x < Width {
				img.Set(int(x), int(y), color)
			}
		}
	}

	interpolateX := func(y, y1, y2, x1, x2 float64) Pixel {
		if y1 == y2 {
			return Pixel(x1)
		}
		return Pixel(x1 + (x2-x1)*(y-y1)/(y2-y1))
	}

	for yf := p1.Y(); yf <= p2.Y(); yf++ {
		x1 := interpolateX(yf, p1.Y(), p2.Y(), p1.X(), p2.X())
		x2 := interpolateX(yf, p1.Y(), p3.Y(), p1.X(), p3.X())
		drawHorizontalLine(Pixel(yf), x1, x2, fillColor)
	}

	for yf := p2.Y(); yf <= p3.Y(); yf++ {
		x1 := interpolateX(yf, p2.Y(), p3.Y(), p2.X(), p3.X())
		x2 := interpolateX(yf, p1.Y(), p3.Y(), p1.X(), p3.X())
		drawHorizontalLine(Pixel(yf), x1, x2, fillColor)
	}
}

func triangleOverlapsScreen(p1, p2, p3 mgl.Vec2, width, height Pixel) bool {
	minX := min(int(p1.X()), min(int(p2.X()), int(p3.X())))
	maxX := max(int(p1.X()), max(int(p2.X()), int(p3.X())))
	minY := min(int(p1.Y()), min(int(p2.Y()), int(p3.Y())))
	maxY := max(int(p1.Y()), max(int(p2.Y()), int(p3.Y())))
	return maxX >= 0 && minX < int(width) && maxY >= 0 && minY < int(height)
}
