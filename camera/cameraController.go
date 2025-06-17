package camera

// Controller is an interface for camera controllers to implement
type Controller interface {
	setCamera(*Camera)
}

// BaseController is a base controller for camera controllers
type BaseController struct {
	camera *Camera
}

// setCamera sets the camera for the controller
func (controller *BaseController) setCamera(camera *Camera) {
	controller.camera = camera
}

// DragController is an interface for Controller that supports dragging
type DragController interface {
	OnDrag(float32, float32)
	OnDragEnd()
}

// ScrollController is an interface for Controller that supports scrolling
type ScrollController interface {
	OnScroll(float32, float32)
}
