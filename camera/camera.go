package camera

import (
	. "ThreeDView/types"
	"github.com/flywave/go-earcut"
	mgl "github.com/go-gl/mathgl/mgl64"
	"log"
	"math"
	"sync"
)

// Camera represents a camera in 3D space
type Camera struct {
	position    mgl.Vec3    // Camera position in world space in units
	fov         Radians     // Field of view in radians
	rotation    mgl.Quat    // Camera rotation as a quaternion
	controller  Controller  // Camera controller
	octree      *octreeNode // Octree for culling
	octreeMutex sync.RWMutex
	widget      ThreeDWidgetInterface

	// Cached values
	viewCache        mgl.Mat4
	perspectiveCache mgl.Mat4
	mvpCache         mgl.Mat4
	frustumCache     Frustum
	aspectRatio      float64
	cacheMutex       sync.RWMutex
}

func (camera *Camera) Position() mgl.Vec3 {
	return camera.position
}

func (camera *Camera) SetPosition(position mgl.Vec3) {
	camera.position = position
}

func (camera *Camera) Fov() Radians {
	return camera.fov
}

func (camera *Camera) SetFov(fov Degrees) {
	camera.fov = fov.ToRadians()
}

func (camera *Camera) Rotation() mgl.Quat {
	return camera.rotation
}

func (camera *Camera) SetRotation(rotation mgl.Quat) {
	camera.rotation = rotation
}

func (camera *Camera) Controller() Controller {
	return camera.controller
}

// SetController sets the controller for the camera. It has to implement the controller interface
func (camera *Camera) SetController(controller Controller) {
	camera.controller = controller
	controller.SetCamera(camera)
}

// NewCamera creates a new camera at the given position in world space and rotation in camera space
func NewCamera(position mgl.Vec3, rotation mgl.Quat, widget ThreeDWidgetInterface) *Camera {
	cam := &Camera{
		position: position,
		rotation: rotation,
		fov:      Degrees(90).ToRadians(),
		widget:   widget,
	}
	cam.UpdateCamera() // Initialize cache
	log.Println("initialized cam cache")
	cam.BuildOctree()
	log.Println("initialized octree")
	widget.SetCamera(cam)
	log.Println("added camera")
	return cam
}

// getFrustumPlanes extracts frustum planes from the camera
func (camera *Camera) getFrustumPlanes() Frustum {
	mvp := camera.mvpCache

	var frustum Frustum
	m := mvp[:]

	// Left plane
	frustum.Planes[0] = Plane{
		Normal: mgl.Vec3{m[3] + m[0], m[7] + m[4], m[11] + m[8]},
		D:      m[15] + m[12],
	}

	// Right plane
	frustum.Planes[1] = Plane{
		Normal: mgl.Vec3{m[3] - m[0], m[7] - m[4], m[11] - m[8]},
		D:      m[15] - m[12],
	}

	// Bottom plane
	frustum.Planes[2] = Plane{
		Normal: mgl.Vec3{m[3] + m[1], m[7] + m[5], m[11] + m[9]},
		D:      m[15] + m[13],
	}

	// Top plane
	frustum.Planes[3] = Plane{
		Normal: mgl.Vec3{m[3] - m[1], m[7] - m[5], m[11] - m[9]},
		D:      m[15] - m[13],
	}

	// Near plane
	frustum.Planes[4] = Plane{
		Normal: mgl.Vec3{m[3] + m[2], m[7] + m[6], m[11] + m[10]},
		D:      m[15] + m[14],
	}

	// Far plane
	frustum.Planes[5] = Plane{
		Normal: mgl.Vec3{m[3] - m[2], m[7] - m[6], m[11] - m[10]},
		D:      m[15] - m[14],
	}

	// Normalize all planes
	for i := range frustum.Planes {
		length := frustum.Planes[i].Normal.Len()
		frustum.Planes[i].Normal = frustum.Planes[i].Normal.Mul(1.0 / length)
		frustum.Planes[i].D /= length
	}

	return frustum
}

// UpdateCamera updates the cameras octree
func (camera *Camera) UpdateCamera() {
	camera.cacheMutex.Lock()
	defer camera.cacheMutex.Unlock()

	// Initialize octree if needed
	if camera.octree == nil {
		bounds := AABB{
			Min: mgl.Vec3{-1000, -1000, -1000},
			Max: mgl.Vec3{1000, 1000, 1000},
		}
		camera.octree = newOctree(bounds, 8, 32)
	}

	// Update cached values
	width, height := camera.widget.GetWidth(), camera.widget.GetHeight()
	camera.aspectRatio = float64(width) / float64(height)
	camera.viewCache = camera.rotation.Mat4().Mul4(mgl.Translate3D(-camera.position.X(), -camera.position.Y(), -camera.position.Z()))
	camera.perspectiveCache = mgl.Perspective(float64(camera.fov), camera.aspectRatio, 0.1, 1e30)
	camera.mvpCache = camera.perspectiveCache.Mul4(camera.viewCache)
	camera.frustumCache = camera.getFrustumPlanes()
}

// GetVisibleFaces returns faces visible in the frustum
func (camera *Camera) GetVisibleFaces() chan FaceData {
	camera.octreeMutex.RLock()
	defer camera.octreeMutex.RUnlock()
	callbackChan := make(chan FaceData, 1000)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		camera.octree.query(camera.frustumCache, callbackChan, &wg)
		wg.Wait()
		close(callbackChan)
	}()
	return callbackChan
}

// Project projects a 3D point to a 2D point on the screen using mgl
func (camera *Camera) Project(point mgl.Vec3) mgl.Vec2 {
	camera.cacheMutex.RLock()
	defer camera.cacheMutex.RUnlock()

	width, height := camera.widget.GetWidth(), camera.widget.GetHeight()
	win := mgl.Project(point, camera.viewCache, camera.perspectiveCache, 0, 0, int(width), int(height))
	return mgl.Vec2{win.X(), float64(height) - win.Y()}
}

// UnProject returns a point at a given distance from the camera along the ray through the screen point
func (camera *Camera) UnProject(point2d mgl.Vec2, distance Unit) mgl.Vec3 {
	camera.cacheMutex.RLock()
	defer camera.cacheMutex.RUnlock()
	width, height := camera.widget.GetWidth(), camera.widget.GetHeight()
	nearPoint, _ := mgl.UnProject(mgl.Vec3{point2d.X(), float64(height) - point2d.Y(), 0.0}, camera.viewCache, camera.perspectiveCache, 0, 0, int(width), int(height))
	farPoint, _ := mgl.UnProject(mgl.Vec3{point2d.X(), float64(height) - point2d.Y(), 1.0}, camera.viewCache, camera.perspectiveCache, 0, 0, int(width), int(height))
	return nearPoint.Add(farPoint.Sub(nearPoint).Normalize().Mul(float64(distance)))
}

// ClipAndProjectFace clips a polygon (in world space) to the camera frustum and returns the resulting polygon(s) in screen space
// If texCoords is provided, texture coordinates will be interpolated for the clipped polygon
func (camera *Camera) ClipAndProjectFace(face FaceData, texCoords ...[3]mgl.Vec2) []struct {
	Points     [3]mgl.Vec2
	Z          [3]float64
	TexCoords  [3]mgl.Vec2
	HasTexture bool
} {
	camera.cacheMutex.RLock()
	mvp := camera.mvpCache
	camera.cacheMutex.RUnlock()
	width, height := camera.widget.GetWidth(), camera.widget.GetHeight()

	vertices := make([]mgl.Vec4, 3)
	for i := 0; i < 3; i++ {
		v := face.Face[i]
		v4 := mgl.Vec4{v.X(), v.Y(), v.Z(), 1}
		vertices[i] = mvp.Mul4x1(v4)
	}

	// Check if texture coordinates are provided
	hasTexture := len(texCoords) > 0
	var texCoordsArray [3]mgl.Vec2
	if hasTexture {
		texCoordsArray = texCoords[0]
	}

	// Create a map to track original vertex indices and their texture coordinates
	vertexToTexCoord := make(map[int]mgl.Vec2)
	for i := 0; i < 3; i++ {
		if hasTexture {
			vertexToTexCoord[i] = texCoordsArray[i]
		}
	}

	// Clip the polygon in homogeneous space
	var clippedVertices []mgl.Vec4
	var clippedTexCoords []mgl.Vec2

	if hasTexture {
		clippedVertices, clippedTexCoords = clipPolygonHomogeneousWithTexCoords(vertices, vertexToTexCoord, hasTexture)
	} else {
		clippedVertices = clipPolygonHomogeneous(vertices)
	}

	if len(clippedVertices) < 3 {
		return nil
	}

	var out2d []mgl.Vec2
	var outz []float64
	var outtex []mgl.Vec2
	for i, v := range clippedVertices {
		if v.W() <= 0 {
			continue
		}
		ndc := v.Mul(1.0 / v.W())
		sx := (ndc.X() + 1) * 0.5 * float64(width)
		sy := (1 - (ndc.Y()+1)*0.5) * float64(height)
		out2d = append(out2d, mgl.Vec2{sx, sy})
		outz = append(outz, ndc.Z()) // Store NDC z value

		if hasTexture {
			outtex = append(outtex, clippedTexCoords[i])
		}
	}
	if len(out2d) < 3 {
		return nil
	}

	var flat []float64
	for _, v := range out2d {
		flat = append(flat, v.X(), v.Y())
	}
	indices, _ := earcut.Earcut(flat, nil, 2)
	var result []struct {
		Points     [3]mgl.Vec2
		Z          [3]float64
		TexCoords  [3]mgl.Vec2
		HasTexture bool
	}
	for i := 0; i+2 < len(indices); i += 3 {
		tri := [3]mgl.Vec2{
			out2d[indices[i]],
			out2d[indices[i+1]],
			out2d[indices[i+2]],
		}
		ztri := [3]float64{
			outz[indices[i]],
			outz[indices[i+1]],
			outz[indices[i+2]],
		}

		var textri [3]mgl.Vec2
		if hasTexture {
			textri = [3]mgl.Vec2{
				outtex[indices[i]],
				outtex[indices[i+1]],
				outtex[indices[i+2]],
			}
		}

		result = append(result, struct {
			Points     [3]mgl.Vec2
			Z          [3]float64
			TexCoords  [3]mgl.Vec2
			HasTexture bool
		}{tri, ztri, textri, hasTexture})
	}
	return result
}

// clipPolygonHomogeneous clips a convex polygon in homogeneous clip space against the canonical view frustum
func clipPolygonHomogeneous(vertices []mgl.Vec4) []mgl.Vec4 {
	planes := [][4]float64{
		{1, 0, 0, 1},  // x <= w
		{-1, 0, 0, 1}, // -x <= w
		{0, 1, 0, 1},  // y <= w
		{0, -1, 0, 1}, // -y <= w
		{0, 0, 1, 1},  // z <= w
		{0, 0, -1, 1}, // -z <= w
	}
	out := vertices
	for _, p := range planes {
		var clipped []mgl.Vec4
		for i := 0; i < len(out); i++ {
			j := (i + 1) % len(out)
			a := out[i]
			b := out[j]
			ad := p[0]*a.X() + p[1]*a.Y() + p[2]*a.Z() + p[3]*a.W()
			bd := p[0]*b.X() + p[1]*b.Y() + p[2]*b.Z() + p[3]*b.W()
			if ad >= 0 {
				clipped = append(clipped, a)
			}
			if (ad >= 0) != (bd >= 0) {
				t := ad / (ad - bd)
				iv := a.Add(b.Sub(a).Mul(t))
				clipped = append(clipped, iv)
			}
		}
		out = clipped
		if len(out) == 0 {
			return nil
		}
	}
	return out
}

// clipPolygonHomogeneousWithTexCoords clips a convex polygon in homogeneous clip space against the canonical view frustum
// and interpolates texture coordinates for any new vertices created during clipping
func clipPolygonHomogeneousWithTexCoords(vertices []mgl.Vec4, texCoords map[int]mgl.Vec2, hasTexture bool) ([]mgl.Vec4, []mgl.Vec2) {
	if !hasTexture {
		return clipPolygonHomogeneous(vertices), nil
	}

	planes := [][4]float64{
		{1, 0, 0, 1},  // x <= w
		{-1, 0, 0, 1}, // -x <= w
		{0, 1, 0, 1},  // y <= w
		{0, -1, 0, 1}, // -y <= w
		{0, 0, 1, 1},  // z <= w
		{0, 0, -1, 1}, // -z <= w
	}

	outVertices := vertices
	outTexCoords := make([]mgl.Vec2, len(vertices))
	for i := 0; i < len(vertices); i++ {
		outTexCoords[i] = texCoords[i]
	}

	for _, p := range planes {
		var clippedVertices []mgl.Vec4
		var clippedTexCoords []mgl.Vec2

		for i := 0; i < len(outVertices); i++ {
			j := (i + 1) % len(outVertices)
			a := outVertices[i]
			b := outVertices[j]
			ta := outTexCoords[i]
			tb := outTexCoords[j]

			ad := p[0]*a.X() + p[1]*a.Y() + p[2]*a.Z() + p[3]*a.W()
			bd := p[0]*b.X() + p[1]*b.Y() + p[2]*b.Z() + p[3]*b.W()

			if ad >= 0 {
				clippedVertices = append(clippedVertices, a)
				clippedTexCoords = append(clippedTexCoords, ta)
			}

			if (ad >= 0) != (bd >= 0) {
				t := ad / (ad - bd)
				iv := a.Add(b.Sub(a).Mul(t))

				// Interpolate texture coordinates
				itex := mgl.Vec2{
					ta.X() + t*(tb.X()-ta.X()),
					ta.Y() + t*(tb.Y()-ta.Y()),
				}

				clippedVertices = append(clippedVertices, iv)
				clippedTexCoords = append(clippedTexCoords, itex)
			}
		}

		outVertices = clippedVertices
		outTexCoords = clippedTexCoords

		if len(outVertices) == 0 {
			return nil, nil
		}
	}

	return outVertices, outTexCoords
}

func (camera *Camera) BuildOctree() {
	camera.octreeMutex.Lock()
	defer camera.octreeMutex.Unlock()
	// Clear existing octree
	bounds := AABB{
		Min: mgl.Vec3{-math.MaxInt, -math.MaxInt, -math.MaxInt},
		Max: mgl.Vec3{math.MaxInt, math.MaxInt, math.MaxInt},
	}
	camera.octree = newOctree(bounds, 8, 32)

	// Get all objects from widget
	objects := camera.widget.GetObjects()

	if len(objects) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(objects))

	for _, obj := range objects {
		go func(obj ObjectInterface) {
			defer wg.Done()
			faces := obj.Faces()
			for _, face := range faces {
				camera.octree.insert(face)
			}
		}(obj)
	}
	wg.Wait()
}
