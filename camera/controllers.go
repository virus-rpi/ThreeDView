package camera

import (
	. "ThreeDView/types"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	mgl "github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
)

type ObjectInterface interface {
	GetPosition() mgl.Vec3
}

// OrbitController is a controller that allows the camera to orbit around a target Object
type OrbitController struct {
	BaseController
	target          ObjectInterface // The Object the camera is orbiting around in world space
	rotation        mgl.Quat        // The rotation of the camera around the target in world space (quaternion)
	distance        Unit            // The distance of the camera from the target
	controlsEnabled bool            // Whether the controls are enabled (dragging, scrolling)
}

// NewOrbitController creates a new OrbitController with the target Object
func NewOrbitController(target ObjectInterface) *OrbitController {
	return &OrbitController{
		target:          target,
		distance:        500,
		rotation:        mgl.QuatIdent(),
		controlsEnabled: true,
	}
}

func (controller *OrbitController) setCamera(camera *Camera) {
	controller.BaseController.camera = camera
	controller.Update()
}

func (controller *OrbitController) SetControlsEnabled(enabled bool) {
	controller.controlsEnabled = enabled
}

func (controller *OrbitController) SetTarget(target ObjectInterface) {
	controller.target = target
	controller.Update()
}

func (controller *OrbitController) SetDistance(distance Unit) {
	controller.distance = distance
	controller.Update()
}

func (controller *OrbitController) Move(distance Unit) {
	controller.distance += distance
	if controller.distance < 1 {
		controller.distance = 1
	}
	controller.Update()
}

// Rotate applies a quaternion delta to the current rotation
func (controller *OrbitController) Rotate(q mgl.Quat) {
	controller.rotation = q.Mul(controller.rotation)
	controller.Update()
}

// SetRotation sets the absolute rotation
func (controller *OrbitController) SetRotation(q mgl.Quat) {
	controller.rotation = q
	controller.Update()
}

// OnDrag rotates the camera around the target using mouse drag (dx, dy in pixels)
func (controller *OrbitController) OnDrag(dx, dy float32) {
	if !controller.controlsEnabled {
		return
	}
	const yawSensitivity = 0.01
	const pitchSensitivity = 0.01
	qYaw := mgl.QuatRotate(float64(dx)*yawSensitivity, mgl.Vec3{0, 1, 0})
	qPitch := mgl.QuatRotate(float64(-dy)*pitchSensitivity, mgl.Vec3{1, 0, 0})
	controller.Rotate(qYaw.Mul(qPitch))
}

func (controller *OrbitController) OnDragEnd() {}

func (controller *OrbitController) OnScroll(_, y float32) {
	if !controller.controlsEnabled {
		return
	}
	controller.Move(Unit(-y * 5))
}

// Update recalculates the camera's position and orientation
func (controller *OrbitController) Update() {
	controller.updatePosition()
	controller.lookAtTarget()
}

func (controller *OrbitController) updatePosition() {
	if controller.camera == nil || controller.target == nil {
		return
	}
	center := controller.target.GetPosition()
	pos := mgl.Vec3{0, 0, float64(controller.distance)}
	pos = controller.rotation.Rotate(pos)
	pos = pos.Add(center)
	controller.camera.Position = pos
}

func (controller *OrbitController) lookAtTarget() {
	if controller.camera == nil || controller.target == nil {
		return
	}
	center := controller.target.GetPosition()
	cameraPos := controller.camera.Position
	direction := center.Sub(cameraPos).Normalize()
	up := mgl.Vec3{0, 0, 1}
	controller.camera.Rotation = mgl.QuatLookAtV(mgl.Vec3{0, 0, 0}, direction, up)
}

// ManualController is a controller that allows the camera to be manually controlled. Useful for debugging
type ManualController struct {
	BaseController
	angles [3]float64 // yaw, pitch, roll in degrees
}

// NewManualController creates a new ManualController
func NewManualController() *ManualController {
	return &ManualController{}
}

// GetRotationSlider returns a container with sliders for controlling the rotation of the camera
func (controller *ManualController) GetRotationSlider() *fyne.Container {
	sliderYaw := widget.NewSlider(0, 360)
	sliderYaw.OnChanged = func(value float64) {
		controller.angles[0] = value
		controller.RefreshCameraRotation()
	}
	sliderPitch := widget.NewSlider(0, 360)
	sliderPitch.OnChanged = func(value float64) {
		controller.angles[1] = value
		controller.RefreshCameraRotation()
	}
	sliderRoll := widget.NewSlider(0, 360)
	sliderRoll.OnChanged = func(value float64) {
		controller.angles[2] = value
		controller.RefreshCameraRotation()
	}
	sliderContainer := container.NewVBox(sliderYaw, sliderPitch, sliderRoll)
	return sliderContainer
}

// GetPositionControl returns a container with sliders for controlling the position of the camera
func (controller *ManualController) GetPositionControl() *fyne.Container {
	sliderX := widget.NewSlider(-100, 100)
	sliderX.OnChanged = func(value float64) {
		if value > 0 {
			controller.camera.Position[0] += 10
		} else {
			controller.camera.Position[0] -= 10
		}
	}
	sliderX.OnChangeEnded = func(value float64) {
		sliderX.Value = 0
	}

	sliderY := widget.NewSlider(-100, 100)
	sliderY.OnChanged = func(value float64) {
		if value > 0 {
			controller.camera.Position[1] += 10
		} else {
			controller.camera.Position[1] -= 10
		}
	}
	sliderY.OnChangeEnded = func(value float64) {
		sliderY.Value = 0
	}

	sliderZ := widget.NewSlider(-100, 100)
	sliderZ.OnChanged = func(value float64) {
		if value > 0 {
			controller.camera.Position[2] += 10
		} else {
			controller.camera.Position[2] -= 10
		}
	}
	sliderZ.OnChangeEnded = func(value float64) {
		sliderZ.Value = 0
	}

	buttonContainer := container.NewVBox(
		sliderX,
		sliderY,
		sliderZ,
	)
	return buttonContainer
}

// GetInfoLabel returns a label that displays the position and rotation of the camera
func (controller *ManualController) GetInfoLabel() *widget.Label {
	label := widget.NewLabel("X: 0 Y: 0 Z: 0      Quaternion: (1, 0, 0, 0)")
	go func() {
		ticker := time.NewTicker(time.Second / 30)
		defer ticker.Stop()
		for range ticker.C {
			q := controller.camera.Rotation
			fyne.Do(func() {
				label.SetText(fmt.Sprintf("X: %.2f Y: %.2f Z: %.2f      Q: (%.2f, %.2f, %.2f, %.2f)",
					controller.camera.Position.X(), controller.camera.Position.Y(), controller.camera.Position.Z(),
					q.W, q.X(), q.Y(), q.Z()))
				label.Refresh()
			})
		}
	}()
	return label
}

func (controller *ManualController) RefreshCameraRotation() {
	if controller.camera == nil {
		return
	}
	yaw := controller.angles[0] * math.Pi / 180
	pitch := controller.angles[1] * math.Pi / 180
	roll := controller.angles[2] * math.Pi / 180
	qYaw := mgl.QuatRotate(yaw, mgl.Vec3{0, 1, 0})
	qPitch := mgl.QuatRotate(pitch, mgl.Vec3{1, 0, 0})
	qRoll := mgl.QuatRotate(roll, mgl.Vec3{0, 0, 1})
	controller.camera.Rotation = qYaw.Mul(qPitch).Mul(qRoll)
}
