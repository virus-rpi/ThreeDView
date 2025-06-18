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
	image                    *canvas.Image // The image that is rendered on
	camera                   *Camera       // The camera of the 3D widget
	objects                  []*Object     // The objects in the 3D widget
	tickMethods              []func()      // The methods that are called every frame
	bgColor                  color.Color   // The background color of the 3D widget
	renderFaceOutlines       bool          // Whether the faces should be rendered with outlines
	renderFaceColors         bool          // Whether the faces should be rendered with colors
	renderEdgeOutline        bool          // Whether to render edge outlines using Z-buffer edge detection
	renderZBufferDebug       bool          // If true, render Z-buffer as grayscale overlay
	fpsCap                   float64       // The maximum frames per second the widget should render at
	tpsCap                   float64       // The maximum ticks per second the widget should tick at
	edgeThreshold            float64       // Base threshold for edge detection
	depthDistanceModulation  float64       // Factor for depth-based modulation of edge threshold
	grazingAngleMaskPower    float64       // Power factor for grazing angle mask
	grazingAngleMaskHardness float64       // Hardness factor for grazing angle mask
}

// NewThreeDWidget creates a new 3D widget
func NewThreeDWidget() *ThreeDWidget {
	w := &ThreeDWidget{
		bgColor:                  color.Transparent,
		renderFaceColors:         true,
		fpsCap:                   math.Inf(1),
		tpsCap:                   math.Inf(1),
		edgeThreshold:            0.05,
		depthDistanceModulation:  0.1,
		grazingAngleMaskPower:    5.0,
		grazingAngleMaskHardness: 0.5,
	}
	w.ExtendBaseWidget(w)
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
		w.image.Image = w.render()
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

// SetRenderEdgeOutline sets whether to render edge outlines using Z-buffer edge detection.
// If true, edges will be detected using the Z-buffer and rendered with a black outline.
func (w *ThreeDWidget) SetRenderEdgeOutline(newVal bool) {
	w.renderEdgeOutline = newVal
}

// SetRenderZBufferDebug sets whether to render the Z-buffer as a grayscale debug overlay.
func (w *ThreeDWidget) SetRenderZBufferDebug(newVal bool) {
	w.renderZBufferDebug = newVal
}

// SetEdgeThreshold sets the base threshold for edge detection.
// Higher values result in fewer edges being detected.
func (w *ThreeDWidget) SetEdgeThreshold(threshold float64) {
	w.edgeThreshold = threshold
}

// SetDepthDistanceModulation sets the factor for depth-based modulation of edge threshold.
// Higher values increase the threshold for distant objects, reducing edge detection in the distance.
func (w *ThreeDWidget) SetDepthDistanceModulation(factor float64) {
	w.depthDistanceModulation = factor
}

// SetGrazingAngleMaskPower sets the power factor for grazing angle mask.
// Higher values make edges at grazing angles more pronounced.
func (w *ThreeDWidget) SetGrazingAngleMaskPower(power float64) {
	w.grazingAngleMaskPower = power
}

// SetGrazingAngleMaskHardness sets the hardness factor for grazing angle mask.
// Higher values make the transition between masked and unmasked edges sharper.
func (w *ThreeDWidget) SetGrazingAngleMaskHardness(hardness float64) {
	w.grazingAngleMaskHardness = hardness
}

func (w *ThreeDWidget) CreateRenderer() fyne.WidgetRenderer {
	return &threeDRenderer{image: w.image}
}

func (w *ThreeDWidget) render() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(Width), int(Height)))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: w.bgColor}, image.Point{}, draw.Src)

	zBuffer := make([][]float64, int(Width))
	for i := range zBuffer {
		zBuffer[i] = make([]float64, Height)
		for j := range zBuffer[i] {
			zBuffer[i][j] = math.Inf(1)
		}
	}

	var faces []FaceData
	var wg3d sync.WaitGroup
	var mu3d sync.Mutex
	wg3d.Add(len(w.objects))
	for _, object := range w.objects {
		go func(object *Object) {
			defer wg3d.Done()
			for _, face := range object.GetFaces() {
				if w.camera.FaceOverlapsFrustum(face.Face, Width, Height) {
					mu3d.Lock()
					faces = append(faces, face)
					mu3d.Unlock()
				}
			}
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
			clippedPolys := w.camera.ClipAndProjectFace(face.Face, Width, Height)
			if clippedPolys == nil {
				return
			}
			for _, tri := range clippedPolys {
				p1, p2, p3 := tri.Points[0], tri.Points[1], tri.Points[2]
				z1, z2, z3 := tri.Z[0], tri.Z[1], tri.Z[2]
				if !triangleOverlapsScreen(p1, p2, p3, Width, Height) {
					continue
				}
				mu.Lock()
				projectedFaces = append(projectedFaces, ProjectedFaceData{
					Face:     [3]mgl.Vec2{p1, p2, p3},
					Z:        [3]float64{z1, z2, z3},
					Color:    face.Color,
					Distance: face.Distance,
				})
				mu.Unlock()
			}
		}(face)
	}
	wg2d.Wait()

	sort.Slice(projectedFaces, func(i, j int) bool {
		return projectedFaces[i].Distance > projectedFaces[j].Distance
	})

	for _, face := range projectedFaces {
		if w.renderFaceColors {
			drawFilledTriangle(img, face.Face, face.Z, face.Color, zBuffer)
		}
	}

	if w.renderFaceOutlines {
		for _, face := range projectedFaces {
			var outlineColor color.Color
			if !w.renderFaceColors {
				outlineColor = face.Color
			} else {
				outlineColor = color.Black
			}
			drawEdge(img, face.Face[0], face.Z[0], face.Face[1], face.Z[1], outlineColor, zBuffer)
			drawEdge(img, face.Face[1], face.Z[1], face.Face[2], face.Z[2], outlineColor, zBuffer)
			drawEdge(img, face.Face[2], face.Z[2], face.Face[0], face.Z[0], outlineColor, zBuffer)
		}
	}

	if w.renderEdgeOutline {
		edgeMask := detectZBufferEdges(zBuffer, w.edgeThreshold, w.depthDistanceModulation, w.grazingAngleMaskPower, w.grazingAngleMaskHardness)
		thickness := 2
		outlineColor := color.Black
		for x := 0; x < int(Width); x++ {
			for y := 0; y < int(Height); y++ {
				if edgeMask[x][y] {
					for dx := -thickness; dx <= thickness; dx++ {
						for dy := -thickness; dy <= thickness; dy++ {
							nx, ny := x+dx, y+dy
							if nx >= 0 && nx < int(Width) && ny >= 0 && ny < int(Height) {
								img.Set(nx, ny, outlineColor)
							}
						}
					}
				}
			}
		}
	}

	if w.renderZBufferDebug {
		debugImg := image.NewGray(img.Bounds())
		var logzVals []float64
		for x := 0; x < int(Width); x++ {
			for y := 0; y < int(Height); y++ {
				z := zBuffer[x][y]
				if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
					logzVals = append(logzVals, math.Log(z))
				}
			}
		}
		if len(logzVals) == 0 {
			return debugImg
		}
		sort.Float64s(logzVals)
		lowIdx := int(0.02 * float64(len(logzVals)))
		highIdx := int(0.98 * float64(len(logzVals)))
		if highIdx <= lowIdx {
			lowIdx = 0
			highIdx = len(logzVals) - 1
		}
		minZ, maxZ := logzVals[lowIdx], logzVals[highIdx]
		if minZ == maxZ {
			minZ, maxZ = 0, 1
		}
		var edgeMask [][]bool
		if w.renderEdgeOutline {
			edgeMask = detectZBufferEdges(zBuffer, w.edgeThreshold, w.depthDistanceModulation, w.grazingAngleMaskPower, w.grazingAngleMaskHardness)
		}
		for x := 0; x < int(Width); x++ {
			for y := 0; y < int(Height); y++ {
				z := zBuffer[x][y]
				var gray uint8 = 255
				if !math.IsInf(z, 1) && z > 0 {
					logz := math.Log(z)
					norm := (logz - minZ) / (maxZ - minZ)
					if norm < 0 {
						norm = 0
					}
					if norm > 1 {
						norm = 1
					}
					gray = uint8(norm * 255)
				}
				if w.renderEdgeOutline && edgeMask != nil && edgeMask[x][y] {
					gray = 0
				}
				debugImg.SetGray(x, y, color.Gray{Y: gray})
			}
		}
		return debugImg
	}
	return img
}

func detectZBufferEdges(zBuffer [][]float64, baseThreshold, depthModulation, grazingAnglePower, grazingAngleHardness float64) [][]bool {
	w, h := len(zBuffer), len(zBuffer[0])
	mask := make([][]bool, w)
	for i := range mask {
		mask[i] = make([]bool, h)
	}
	var logzVals []float64
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			z := zBuffer[x][y]
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				logzVals = append(logzVals, math.Log(z))
			}
		}
	}
	if len(logzVals) == 0 {
		return mask
	}
	sort.Float64s(logzVals)
	lowIdx := int(0.02 * float64(len(logzVals)))
	highIdx := int(0.98 * float64(len(logzVals)))
	if highIdx <= lowIdx {
		lowIdx = 0
		highIdx = len(logzVals) - 1
	}
	minZ, maxZ := logzVals[lowIdx], logzVals[highIdx]
	if minZ == maxZ {
		minZ, maxZ = 0, 1
	}
	normZ := make([][]float64, w)
	for x := 0; x < w; x++ {
		normZ[x] = make([]float64, h)
		for y := 0; y < h; y++ {
			z := zBuffer[x][y]
			logz := maxZ
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				logz = math.Log(z)
			}
			normZ[x][y] = (logz - minZ) / (maxZ - minZ)
		}
	}
	gx := [3][3]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	gy := [3][3]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}
	for x := 1; x < w-1; x++ {
		for y := 1; y < h-1; y++ {
			var sx, sy float64
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					z := normZ[x+i][y+j]
					sx += gx[i+1][j+1] * z
					sy += gy[i+1][j+1] * z
				}
			}

			z := zBuffer[x][y]
			threshold := baseThreshold
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				threshold *= 1.0 + depthModulation*z
			}

			gradientMagnitude := math.Hypot(sx, sy)
			if gradientMagnitude > 0.001 {
				nx, ny := -sx/gradientMagnitude, -sy/gradientMagnitude

				nz := math.Sqrt(1.0 - nx*nx - ny*ny)
				if math.IsNaN(nz) {
					nz = 0.0
				}
				dotProduct := nz

				fresnel := math.Pow(1.0-math.Max(0.0, math.Min(1.0, dotProduct)), grazingAnglePower)
				grazingAngleMask := math.Max(0.0, math.Min(1.0, (fresnel+grazingAngleHardness-1.0)/grazingAngleHardness))
				threshold *= 1.0 / (1.0 + grazingAngleMask)
			}

			if gradientMagnitude > threshold {
				mask[x][y] = true
			}
		}
	}
	return mask
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

func triangleOverlapsScreen(p1, p2, p3 mgl.Vec2, width, height Pixel) bool {
	minX := min(int(p1.X()), min(int(p2.X()), int(p3.X())))
	maxX := max(int(p1.X()), max(int(p2.X()), int(p3.X())))
	minY := min(int(p1.Y()), min(int(p2.Y()), int(p3.Y())))
	maxY := max(int(p1.Y()), max(int(p2.Y()), int(p3.Y())))
	return maxX >= 0 && minX < int(width) && maxY >= 0 && minY < int(height)
}

func drawFilledTriangle(img *image.RGBA, p [3]mgl.Vec2, z [3]float64, fill color.Color, zBuffer [][]float64) {
	v := [3]struct {
		p mgl.Vec2
		z float64
	}{{p[0], z[0]}, {p[1], z[1]}, {p[2], z[2]}}
	if v[1].p.Y() < v[0].p.Y() {
		v[0], v[1] = v[1], v[0]
	}
	if v[2].p.Y() < v[0].p.Y() {
		v[0], v[2] = v[2], v[0]
	}
	if v[2].p.Y() < v[1].p.Y() {
		v[1], v[2] = v[2], v[1]
	}
	interpolate := func(y, y1, y2, x1, x2, z1, z2 float64) (Pixel, float64) {
		if y1 == y2 {
			return Pixel(x1), z1
		}
		t := (y - y1) / (y2 - y1)
		return Pixel(x1 + (x2-x1)*t), z1 + (z2-z1)*t
	}
	for yf := math.Ceil(v[0].p.Y()); yf <= v[1].p.Y(); yf++ {
		x1, z1 := interpolate(yf, v[0].p.Y(), v[1].p.Y(), v[0].p.X(), v[1].p.X(), v[0].z, v[1].z)
		x2, z2 := interpolate(yf, v[0].p.Y(), v[2].p.Y(), v[0].p.X(), v[2].p.X(), v[0].z, v[2].z)
		if x1 > x2 {
			x1, x2, z1, z2 = x2, x1, z2, z1
		}
		for x := int(math.Ceil(float64(x1))); float64(x) <= float64(x2); x++ {
			if x >= 0 && x < int(Width) && int(yf) >= 0 && int(yf) < int(Height) {
				t := 0.0
				if x2 != x1 {
					t = float64(x-int(x1)) / float64(x2-x1)
				}
				z := z1 + (z2-z1)*t
				if z < zBuffer[x][int(yf)] {
					zBuffer[x][int(yf)] = z
					img.Set(x, int(yf), fill)
				}
			}
		}
	}
	for yf := v[1].p.Y(); yf <= v[2].p.Y(); yf++ {
		x1, z1 := interpolate(yf, v[1].p.Y(), v[2].p.Y(), v[1].p.X(), v[2].p.X(), v[1].z, v[2].z)
		x2, z2 := interpolate(yf, v[0].p.Y(), v[2].p.Y(), v[0].p.X(), v[2].p.X(), v[0].z, v[2].z)
		if x1 > x2 {
			x1, x2, z1, z2 = x2, x1, z2, z1
		}
		for x := int(math.Ceil(float64(x1))); float64(x) <= float64(x2); x++ {
			if x >= 0 && x < int(Width) && int(yf) >= 0 && int(yf) < int(Height) {
				t := 0.0
				if x2 != x1 {
					t = float64(x-int(x1)) / float64(x2-x1)
				}
				z := z1 + (z2-z1)*t
				if z < zBuffer[x][int(yf)] {
					zBuffer[x][int(yf)] = z
					img.Set(x, int(yf), fill)
				}
			}
		}
	}
}

func drawEdge(img *image.RGBA, p1 mgl.Vec2, z1 float64, p2 mgl.Vec2, z2 float64, c color.Color, zBuffer [][]float64) {
	x0, y0 := int(math.Round(p1.X())), int(math.Round(p1.Y()))
	x1, y1 := int(math.Round(p2.X())), int(math.Round(p2.Y()))
	dx, dy := int(math.Abs(float64(x1-x0))), int(math.Abs(float64(y1-y0)))
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err, maxIter, iter := dx-dy, dx+dy+10, 0
	for {
		if x0 >= 0 && x0 < int(Width) && y0 >= 0 && y0 < int(Height) {
			t := 0.0
			total := math.Hypot(float64(x1-x0), float64(y1-y0))
			if total != 0 {
				t = math.Hypot(float64(x0-int(math.Round(p1.X()))), float64(y0-int(math.Round(p1.Y())))) / total
			}
			z := z1 + (z2-z1)*t
			if z <= zBuffer[x0][y0] {
				img.Set(x0, y0, c)
			}
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
			break
		}
	}
}
