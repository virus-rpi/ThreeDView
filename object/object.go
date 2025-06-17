package object

import (
	. "ThreeDView/camera"
	. "ThreeDView/types"
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
	Face     [3]Point2D  // The Face in 2D space as 3 2d points
	Color    color.Color // The Color of the Face
	Distance Unit        // The Distance of the un-projected Face from the camera in 3d world space
}

// Object represents a 3D shape in world space
type Object struct {
	Faces    []FaceData            // Faces of the Object in local space
	Rotation Quaternion            // Rotation of the Object in world space (now quaternion)
	Position Point3D               // Position of the Object in world space
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
			face.Face.Rotate(Point3D{}, object.Rotation)
			face.Face.Add(object.Position)
			face.Distance = face.Face.DistanceTo(object.Widget.GetCamera().Position)
			faces[i] = face
		}(i, face)
	}
	wg.Wait()
	return faces
}

func (object *Object) GetPosition() Point3D {
	return object.Position
}

func (object *Object) GetRotation() Quaternion {
	return object.Rotation
}
