package ThreeDView

import (
	. "ThreeDView/camera"
	. "ThreeDView/object"
	"ThreeDView/renderer"
	. "ThreeDView/types"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image/color"
	"log"
	"math"
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
	image               *canvas.Image // The image that is rendered on
	camera              *Camera       // The camera of the 3D widget
	objects             []*Object     // The objects in the 3D widget
	tickMethods         []func()      // The methods that are called every frame
	bgColor             color.Color   // The background color of the 3D widget
	renderFaceOutlines  bool          // Whether the faces should be rendered with outlines
	renderFaceColors    bool          // Whether the faces should be rendered with colors
	renderTextures      bool          // Whether to use textures for rendering (if available)
	renderEdgeOutline   bool          // Whether to render edge outlines using Z-buffer edge detection
	renderZBuffer       bool          // If true, render Z-buffer as grayscale overlay
	renderPseudoShading bool          // If true, render pseudo-shading based on depth
	fpsCap              float64       // The maximum frames per second the widget should render at
	tpsCap              float64       // The maximum ticks per second the widget should tick at
	renderer            *renderer.Renderer
}

// NewThreeDWidget creates a new 3D widget
func NewThreeDWidget() *ThreeDWidget {
	w := &ThreeDWidget{
		bgColor:             color.Transparent,
		renderFaceColors:    true,
		renderTextures:      true,
		renderPseudoShading: true,
		fpsCap:              math.Inf(1),
		tpsCap:              math.Inf(1),
	}
	w.renderer = renderer.NewRenderer(w)
	w.ExtendBaseWidget(w)
	standardCamera := NewCamera(mgl.Vec3{}, mgl.QuatIdent())
	w.camera = &standardCamera
	w.objects = []*Object{}
	w.image = canvas.NewImageFromImage(w.renderer.Render())
	go w.renderLoop()
	go w.tickLoop()
	return w
}

func (w *ThreeDWidget) tickLoop() {
	for {
		if w.tpsCap == 0 || !w.Visible() {
			continue
		}
		start := time.Now()
		tickDur := time.Second / time.Duration(w.tpsCap)
		for _, tick := range w.tickMethods {
			tick()
		}
		elapsed := time.Since(start)
		if elapsed < tickDur {
			time.Sleep(tickDur - elapsed)
		}
		if elapsed > tickDur && tickDur != 0 {
			log.Println("WARNING: Tick took too long (", elapsed, " > ", tickDur, ")")
		}
	}
}

func (w *ThreeDWidget) renderLoop() {
	for {
		if w.fpsCap == 0 || !w.Visible() {
			continue
		}
		start := time.Now()
		frameDur := time.Second / time.Duration(w.fpsCap)
		w.image.Image = w.renderer.Render()
		fyne.Do(func() { canvas.Refresh(w.image) })
		elapsed := time.Since(start)
		if elapsed < frameDur {
			time.Sleep(frameDur - elapsed)
		}
		if elapsed > frameDur && frameDur != 0 {
			log.Println("WARNING: Frame took too long (", elapsed, " > ", frameDur, ")")
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

func (w *ThreeDWidget) GetBackgroundColor() color.Color { return w.bgColor }

func (w *ThreeDWidget) GetObjects() []*Object { return w.objects }

func (w *ThreeDWidget) GetRenderFaceColors() bool {
	return w.renderFaceColors
}

func (w *ThreeDWidget) GetRenderTextures() bool {
	return w.renderTextures
}

func (w *ThreeDWidget) GetRenderFaceOutlines() bool {
	return w.renderFaceOutlines
}

func (w *ThreeDWidget) GetRenderEdgeOutlines() bool {
	return w.renderEdgeOutline
}

func (w *ThreeDWidget) GetRenderZBuffer() bool {
	return w.renderZBuffer
}

func (w *ThreeDWidget) GetRenderPseudoShading() bool {
	return w.renderPseudoShading
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

// SetRenderTextures sets whether textures should be used for rendering (if available).
// If true, faces with texture information will be rendered using their texture.
// If false, all faces will be rendered using their solid color.
// Default is true
func (w *ThreeDWidget) SetRenderTextures(newVal bool) {
	w.renderTextures = newVal
}

// SetRenderEdgeOutline sets whether to render edge outlines using Z-buffer edge detection.
// If true, edges will be detected using the Z-buffer and rendered with a black outline.
func (w *ThreeDWidget) SetRenderEdgeOutline(newVal bool) {
	w.renderEdgeOutline = newVal
}

// SetRenderZBufferDebug sets whether to render the Z-buffer as a grayscale debug overlay.
func (w *ThreeDWidget) SetRenderZBufferDebug(newVal bool) {
	w.renderZBuffer = newVal
}

func (w *ThreeDWidget) SetRenderPseudoShading(newVal bool) {
	w.renderPseudoShading = newVal
}

func (w *ThreeDWidget) CreateRenderer() fyne.WidgetRenderer {
	return &threeDRenderer{image: w.image}
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

type threeDRenderer struct{ image *canvas.Image }

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
