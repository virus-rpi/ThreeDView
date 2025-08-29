package camera

import "github.com/virus-rpi/ThreeDView/types"

// BaseController is a base controller for camera controllers
type BaseController struct {
	camera types.CameraInterface
}

// SetCamera sets the camera for the controller
func (controller *BaseController) SetCamera(camera types.CameraInterface) {
	controller.camera = camera
}

// DragController is an interface for controller that supports dragging
type DragController interface {
	OnDrag(float32, float32)
	OnDragEnd()
}

// ScrollController is an interface for controller that supports scrolling
type ScrollController interface {
	OnScroll(float32, float32)
}
