package object

import (
	mgl "github.com/go-gl/mathgl/mgl64"
	"github.com/virus-rpi/ThreeDView/types"
	"image"
	"image/color"
	"sync"
)

// ProjectedFaceData represents a face projected to 2D space
type ProjectedFaceData struct {
	Face         [3]mgl.Vec2 // The Face in 2D space as 3 2d points
	Z            [3]float64  // The Z (depth) value for each vertex
	Color        color.Color // The Color of the Face
	Distance     types.Unit  // The Distance of the un-projected Face from the camera in 3d world space
	TextureImage image.Image // The texture image for the face (nil if no texture)
	TexCoords    [3]mgl.Vec2 // Texture coordinates for each vertex
	HasTexture   bool        // Whether this face has texture information
}

// Object represents a 3D shape in world space
type Object struct {
	faces    []types.FaceData            // faces of the Object in local space
	rotation mgl.Quat                    // Rotation of the Object in world space (now quaternion)
	position mgl.Vec3                    // Position of the Object in world space
	widget   types.ThreeDWidgetInterface // The widget the Object is in
}

func (object *Object) SetFaces(faces []types.FaceData) {
	object.faces = faces
	object.widget.GetCamera().RebuildOctree()
}

func (object *Object) Rotation() mgl.Quat {
	return object.rotation
}

func (object *Object) SetRotation(rotation mgl.Quat) {
	object.rotation = rotation
	object.widget.GetCamera().RebuildOctree()
}

func (object *Object) Position() mgl.Vec3 {
	return object.position
}

func (object *Object) SetPosition(position mgl.Vec3) {
	object.position = position
	object.widget.GetCamera().RebuildOctree()
}

func (object *Object) Widget() types.ThreeDWidgetInterface {
	return object.widget
}

func (object *Object) SetWidget(widget types.ThreeDWidgetInterface) {
	object.widget = widget
	object.widget.GetCamera().RebuildOctree()
}

func (object *Object) transformFace(face types.FaceData) types.FaceData {
	clonedFace := face
	clonedFace.Face = face.Face
	clonedFace.Rotate(mgl.Vec3{}, object.rotation)
	clonedFace.Add(object.position)
	clonedFace.Distance = clonedFace.DistanceTo(object.widget.GetCamera().Position())

	clonedFace.TextureImage = face.TextureImage
	clonedFace.TexCoords = face.TexCoords
	clonedFace.HasTexture = face.HasTexture

	return clonedFace
}

func (object *Object) Faces() []types.FaceData {
	faces := make([]types.FaceData, len(object.faces))
	var wg sync.WaitGroup
	wg.Add(len(object.faces))

	for i, face := range object.faces {
		go func(i int, face types.FaceData) {
			defer wg.Done()
			faces[i] = object.transformFace(face)
		}(i, face)
	}

	wg.Wait()
	return faces
}

func (object *Object) StreamFaces() <-chan types.FaceData {
	out := make(chan types.FaceData, 1000)

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(object.faces))

		for _, face := range object.faces {
			go func(face types.FaceData) {
				defer wg.Done()
				out <- object.transformFace(face)
			}(face)
		}

		wg.Wait()
		close(out)
	}()

	return out
}
