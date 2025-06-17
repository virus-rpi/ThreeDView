package camera

import (
	. "ThreeDView/types"
	mgl "github.com/go-gl/mathgl/mgl64"
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
func (camera *Camera) Project(point mgl.Vec3, width, height Pixel) Point2D {
	view := camera.Rotation.Mat4().Mul4(mgl.Translate3D(-camera.Position.X(), -camera.Position.Y(), -camera.Position.Z()))
	aspect := float64(width) / float64(height)
	proj := mgl.Perspective(float64(camera.Fov.ToRadians()), aspect, 0.1, 10000.0)
	win := mgl.Project(point, view, proj, 0, 0, int(width), int(height))
	return Point2D{X: Pixel(win.X()), Y: Pixel(float64(height) - win.Y())}
}

// UnProject returns a point at a given distance from the camera along the ray through the screen point
func (camera *Camera) UnProject(point2d Point2D, distance Unit, width, height Pixel) mgl.Vec3 {
	winNear := mgl.Vec3{float64(point2d.X), float64(height) - float64(point2d.Y), 0.0}
	winFar := mgl.Vec3{float64(point2d.X), float64(height) - float64(point2d.Y), 1.0}
	view := camera.Rotation.Mat4().Mul4(mgl.Translate3D(-camera.Position.X(), -camera.Position.Y(), -camera.Position.Z()))
	aspect := float64(width) / float64(height)
	proj := mgl.Perspective(float64(camera.Fov.ToRadians()), aspect, 0.1, 10000.0)
	nearPoint, _ := mgl.UnProject(winNear, view, proj, 0, 0, int(width), int(height))
	farPoint, _ := mgl.UnProject(winFar, view, proj, 0, 0, int(width), int(height))
	dir := farPoint.Sub(nearPoint).Normalize()
	return nearPoint.Add(dir.Mul(float64(distance)))
}

// FaceOverlapsFrustum returns true if any part of the face is inside the camera frustum
func (camera *Camera) FaceOverlapsFrustum(face Face, width, height Pixel) bool {
	view := camera.Rotation.Mat4().Mul4(mgl.Translate3D(-camera.Position.X(), -camera.Position.Y(), -camera.Position.Z()))
	aspect := float64(width) / float64(height)
	proj := mgl.Perspective(float64(camera.Fov.ToRadians()), aspect, 0.1, 10000.0)

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

	center := mgl.Vec2{float64(width) / 2, float64(height) / 2}
	if pointInTriangle(center, mgl.Vec2{projected[0].X(), projected[0].Y()}, mgl.Vec2{projected[1].X(), projected[1].Y()}, mgl.Vec2{projected[2].X(), projected[2].Y()}) {
		return true
	}

	return false
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
