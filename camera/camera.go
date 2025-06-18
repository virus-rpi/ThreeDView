package camera

import (
	. "ThreeDView/types"
	"github.com/flywave/go-earcut"
	mgl "github.com/go-gl/mathgl/mgl64"
	"math"
)

// Camera represents a camera in 3D space
type Camera struct {
	Position   mgl.Vec3   // Camera position in world space in units
	Fov        Degrees    // Field of view in degrees
	Rotation   mgl.Quat   // Camera rotation as a quaternion
	Controller Controller // Camera Controller
}

// NewCamera creates a new camera at the given position in world space and rotation in camera space
func NewCamera(position mgl.Vec3, rotation mgl.Quat) Camera {
	return Camera{Position: position, Rotation: rotation, Fov: 90}
}

// SetController sets the controller for the camera. It has to implement the Controller interface
func (camera *Camera) SetController(controller Controller) {
	camera.Controller = controller
	controller.setCamera(camera)
}

// Project projects a 3D point to a 2D point on the screen using mgl
func (camera *Camera) Project(point mgl.Vec3, width, height Pixel) mgl.Vec2 {
	view := camera.Rotation.Mat4().Mul4(mgl.Translate3D(-camera.Position.X(), -camera.Position.Y(), -camera.Position.Z()))
	aspect := float64(width) / float64(height)
	proj := mgl.Perspective(float64(camera.Fov.ToRadians()), aspect, 0.1, math.MaxFloat64)
	win := mgl.Project(point, view, proj, 0, 0, int(width), int(height))
	return mgl.Vec2{win.X(), float64(height) - win.Y()}
}

// UnProject returns a point at a given distance from the camera along the ray through the screen point
func (camera *Camera) UnProject(point2d mgl.Vec2, distance Unit, width, height Pixel) mgl.Vec3 {
	winNear := mgl.Vec3{point2d.X(), float64(height) - point2d.Y(), 0.0}
	winFar := mgl.Vec3{point2d.X(), float64(height) - point2d.Y(), 1.0}
	view := camera.Rotation.Mat4().Mul4(mgl.Translate3D(-camera.Position.X(), -camera.Position.Y(), -camera.Position.Z()))
	aspect := float64(width) / float64(height)
	proj := mgl.Perspective(float64(camera.Fov.ToRadians()), aspect, 0.1, math.MaxFloat64)
	nearPoint, _ := mgl.UnProject(winNear, view, proj, 0, 0, int(width), int(height))
	farPoint, _ := mgl.UnProject(winFar, view, proj, 0, 0, int(width), int(height))
	dir := farPoint.Sub(nearPoint).Normalize()
	return nearPoint.Add(dir.Mul(float64(distance)))
}

// FaceOverlapsFrustum returns true if any part of the face is inside the camera frustum
func (camera *Camera) FaceOverlapsFrustum(face Face, width, height Pixel) bool {
	view := camera.Rotation.Mat4().Mul4(mgl.Translate3D(-camera.Position.X(), -camera.Position.Y(), -camera.Position.Z()))
	aspect := float64(width) / float64(height)
	proj := mgl.Perspective(float64(camera.Fov.ToRadians()), aspect, 0.1, math.MaxFloat64)

	projected := [3]mgl.Vec3{}
	for i := 0; i < 3; i++ {
		projected[i] = mgl.Project(face[i], view, proj, 0, 0, int(width), int(height))
	}

	for i := 0; i < 3; i++ {
		x, y := projected[i].X(), projected[i].Y()
		if x >= 0 && x < float64(width) && y >= 0 && y < float64(height) {
			return true
		}
	}

	testEdge := func(p1, p2 mgl.Vec3) bool {
		screenEdges := [][2]mgl.Vec2{
			{{0, 0}, {float64(width), 0}},
			{{float64(width), 0}, {float64(width), float64(height)}},
			{{float64(width), float64(height)}, {0, float64(height)}},
			{{0, float64(height)}, {0, 0}},
		}
		for _, edge := range screenEdges {
			if linesIntersect(
				mgl.Vec2{p1.X(), p1.Y()}, mgl.Vec2{p2.X(), p2.Y()},
				edge[0], edge[1],
			) {
				return true
			}
		}
		return false
	}

	for i := 0; i < 3; i++ {
		if testEdge(projected[i], projected[(i+1)%3]) {
			return true
		}
	}

	corners := []mgl.Vec2{
		{0, 0},
		{float64(width), 0},
		{float64(width), float64(height)},
		{0, float64(height)},
		{float64(width) / 2, float64(height) / 2}, // center
	}
	tri := [3]mgl.Vec2{
		{projected[0].X(), projected[0].Y()},
		{projected[1].X(), projected[1].Y()},
		{projected[2].X(), projected[2].Y()},
	}
	for _, corner := range corners {
		if pointInTriangle(corner, tri[0], tri[1], tri[2]) {
			return true
		}
	}

	return false
}

// ClipAndProjectFace clips a polygon (in world space) to the camera frustum and returns the resulting polygon(s) in screen space
// If texCoords is provided, texture coordinates will be interpolated for the clipped polygon
func (camera *Camera) ClipAndProjectFace(face Face, width, height Pixel, texCoords ...[3]mgl.Vec2) []struct {
	Points     [3]mgl.Vec2
	Z          [3]float64
	TexCoords  [3]mgl.Vec2
	HasTexture bool
} {
	view := camera.Rotation.Mat4().Mul4(mgl.Translate3D(-camera.Position.X(), -camera.Position.Y(), -camera.Position.Z()))
	aspect := float64(width) / float64(height)

	// Use a very large far plane, but not math.MaxFloat64 to avoid numerical instability
	farPlane := 1e30 // Much larger than before, but not so large it causes precision issues
	proj := mgl.Perspective(float64(camera.Fov.ToRadians()), aspect, 0.1, farPlane)
	mvp := proj.Mul4(view)

	vertices := make([]mgl.Vec4, 3)
	for i := 0; i < 3; i++ {
		v := face[i]
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
		if v.W() <= 0 { // Changed from == 0 to <= 0 to handle points at or beyond infinity
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

func linesIntersect(p1, p2, q1, q2 mgl.Vec2) bool {
	ccw := func(a, b, c mgl.Vec2) bool {
		return (c.Y()-a.Y())*(b.X()-a.X()) > (b.Y()-a.Y())*(c.X()-a.X())
	}
	return (ccw(p1, q1, q2) != ccw(p2, q1, q2)) && (ccw(p1, p2, q1) != ccw(p1, p2, q2))
}

func pointInTriangle(p, a, b, c mgl.Vec2) bool {
	v0 := c.Sub(a)
	v1 := b.Sub(a)
	v2 := p.Sub(a)
	dot00 := v0.Dot(v0)
	dot01 := v0.Dot(v1)
	dot02 := v0.Dot(v2)
	dot11 := v1.Dot(v1)
	dot12 := v1.Dot(v2)
	invDenom := 1 / (dot00*dot11 - dot01*dot01)
	u := (dot11*dot02 - dot01*dot12) * invDenom
	v := (dot00*dot12 - dot01*dot02) * invDenom
	return (u >= 0) && (v >= 0) && (u+v <= 1)
}
