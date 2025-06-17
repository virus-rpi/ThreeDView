package camera

import (
	. "ThreeDView/types"
	"math"
)

// Camera represents a camera in 3D space
type Camera struct {
	Position   Point3D    // Camera position in world space in units
	Fov        Degrees    // Field of view in degrees
	Rotation   Quaternion // Camera rotation as a quaternion
	Controller Controller // Camera Controller
}

// NewCamera creates a new camera at the given position in world space and rotation in camera space
func NewCamera(position Point3D, rotation Quaternion) Camera {
	return Camera{Position: position, Rotation: rotation, Fov: 90}
}

// SetController sets the controller for the camera. It has to implement the Controller interface
func (camera *Camera) SetController(controller Controller) {
	camera.Controller = controller
	controller.setCamera(camera)
}

// Project projects a 3D point to a 2D point on the screen
func (camera *Camera) Project(point Point3D, width, height Pixel) Point2D {
	translatedPoint := point
	translatedPoint.Subtract(camera.Position)
	translatedPoint = camera.Rotation.RotatePoint(translatedPoint)

	epsilon := Unit(0.0001)
	if math.Abs(float64(translatedPoint.Z)) < float64(epsilon) {
		translatedPoint.Z = epsilon
	}

	fovRadians := camera.Fov.ToRadians()
	scale := Unit(float64(width) / (2 * math.Tan(float64(fovRadians/2))))

	x2D := (translatedPoint.X * scale / translatedPoint.Z) + Unit(width)/2
	y2D := (translatedPoint.Y * scale / translatedPoint.Z) + Unit(height)/2

	return Point2D{X: Pixel(x2D), Y: Pixel(y2D)}
}

// UnProject un-projects a 2D point on the screen to a 3D point in world space
func (camera *Camera) UnProject(point2d Point2D, distance Unit, width, height Pixel) Point3D {
	fovRadians := camera.Fov.ToRadians()
	halfWidth := float64(width) / 2
	halfHeight := float64(height) / 2
	scale := math.Tan(float64(fovRadians)/2) * float64(distance)

	pointInCameraSpace := Point3D{
		X: Unit((float64(point2d.X) - halfWidth) / halfWidth * scale),
		Y: Unit((float64(point2d.Y) - halfHeight) / halfHeight * scale),
		Z: distance,
	}

	// Inverse rotate using quaternion
	inv := camera.Rotation.Conjugate()
	pointInWorldSpace := inv.RotatePoint(pointInCameraSpace)
	pointInWorldSpace.Add(camera.Position)

	return pointInWorldSpace
}

func (camera *Camera) FaceOverlapsFrustum(face Face) bool {
	points := [3]Point3D{}
	for i := 0; i < 3; i++ {
		p := face[i]
		p.Subtract(camera.Position)
		p = camera.Rotation.RotatePoint(p)
		points[i] = p
	}

	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	minZ, maxZ := points[0].Z, points[0].Z
	for i := 1; i < 3; i++ {
		if points[i].X < minX {
			minX = points[i].X
		}
		if points[i].X > maxX {
			maxX = points[i].X
		}
		if points[i].Y < minY {
			minY = points[i].Y
		}
		if points[i].Y > maxY {
			maxY = points[i].Y
		}
		if points[i].Z < minZ {
			minZ = points[i].Z
		}
		if points[i].Z > maxZ {
			maxZ = points[i].Z
		}
	}

	fovRadians := camera.Fov.ToRadians()
	aspectRatio := 1.0
	tanFovOver2 := math.Tan(float64(fovRadians) / 2)
	near := Unit(0.1)

	if maxZ < near {
		return false
	}
	rightPlaneX := maxZ * Unit(tanFovOver2*aspectRatio)
	if minX > rightPlaneX || maxX < -rightPlaneX {
		return false
	}
	topPlaneY := maxZ * Unit(tanFovOver2)
	if minY > topPlaneY || maxY < -topPlaneY {
		return false
	}
	return true
}
