package object

import (
	. "ThreeDView/camera"
	. "ThreeDView/types"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image/color"
	"sync"
)

type ThreeDWidgetInterface interface {
	GetCamera() *Camera
	RegisterTickMethod(func())
	AddObject(*Object)
	GetWidth() Pixel
	GetHeight() Pixel
}

// FaceData represents a face in 3D space
type FaceData struct {
	Face     Face        // The Face in 3D space as a Face
	Color    color.Color // The Color of the Face
	Distance Unit        // The Distance of the Face from the camera 3d world space
}

// ProjectedFaceData represents a face projected to 2D space
type ProjectedFaceData struct {
	Face     [3]mgl.Vec2 // The Face in 2D space as 3 2d points
	Z        [3]float64  // The Z (depth) value for each vertex
	Color    color.Color // The Color of the Face
	Distance Unit        // The Distance of the un-projected Face from the camera in 3d world space
}

// Object represents a 3D shape in world space
type Object struct {
	Faces    []FaceData            // Faces of the Object in local space
	Rotation mgl.Quat              // Rotation of the Object in world space (now quaternion)
	Position mgl.Vec3              // Position of the Object in world space
	Widget   ThreeDWidgetInterface // The Widget the Object is in
}

// GetFaces returns the faces of the shape in world space as FaceData
func (object *Object) GetFaces() []FaceData {
	faces := make([]FaceData, len(object.Faces))
	var wg sync.WaitGroup
	wg.Add(len(object.Faces))
	for i, face := range object.Faces {
		go func(i int, face FaceData) {
			defer wg.Done()
			clonedFace := face
			clonedFace.Face = face.Face
			clonedFace.Face.Rotate(mgl.Vec3{}, object.Rotation)
			clonedFace.Face.Add(object.Position)
			clonedFace.Distance = clonedFace.Face.DistanceTo(object.Widget.GetCamera().Position)
			faces[i] = clonedFace
		}(i, face)
	}
	wg.Wait()
	return faces
}

func (object *Object) GetPosition() mgl.Vec3 {
	return object.Position
}

func (object *Object) GetRotation() mgl.Quat {
	return object.Rotation
}
