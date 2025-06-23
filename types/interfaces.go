package types

import (
	mgl "github.com/go-gl/mathgl/mgl64"
	"image/color"
)

type ObjectInterface interface {
	Faces() []FaceData
	Position() mgl.Vec3
	Rotation() mgl.Quat
	Widget() ThreeDWidgetInterface
	SetFaces(faces []FaceData)
	SetPosition(position mgl.Vec3)
	SetRotation(rotation mgl.Quat)
	SetWidget(widget ThreeDWidgetInterface)
}

type Controller interface {
	SetCamera(cam CameraInterface)
}

type CameraInterface interface {
	GetVisibleFaces() []FaceData
	ClipAndProjectFace(face FaceData, texCoords ...[3]mgl.Vec2) []struct {
		Points     [3]mgl.Vec2
		Z          [3]float64
		TexCoords  [3]mgl.Vec2
		HasTexture bool
	}
	UnProject(point2d mgl.Vec2, distance Unit) mgl.Vec3
	BuildOctree()
	UpdateCamera()
	Controller() Controller
	SetController(controller Controller)
	Position() mgl.Vec3
	SetPosition(position mgl.Vec3)
	Rotation() mgl.Quat
	SetRotation(rotation mgl.Quat)
	Fov() Radians
	SetFov(fov Degrees)
}

type ThreeDWidgetInterface interface {
	RegisterTickMethod(func())
	GetWidth() Pixel
	GetHeight() Pixel
	GetBackgroundColor() color.Color
	GetRenderFaceColors() bool
	GetRenderTextures() bool
	GetRenderFaceOutlines() bool
	GetRenderEdgeOutlines() bool
	GetRenderZBuffer() bool
	GetRenderPseudoShading() bool
	GetObjects() []ObjectInterface
	AddObject(obj ObjectInterface)
	SetCamera(camera CameraInterface)
	GetCamera() CameraInterface
}
