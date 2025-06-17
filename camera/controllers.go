package camera

import (
	. "ThreeDView/types"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"math"
	"time"
)

type ObjectInterface interface {
	GetPosition() Point3D
}

// OrbitController is a controller that allows the camera to orbit around a target Object
type OrbitController struct {
	BaseController
	target          ObjectInterface // The Object the camera is orbiting around in world space
	rotation        Quaternion      // The rotation of the camera around the target in world space (quaternion)
	distance        Unit            // The distance of the camera from the target
	controlsEnabled bool            // Whether the controls are enabled (dragging, scrolling)
}

// NewOrbitController creates a new OrbitController with the target Object
func NewOrbitController(target ObjectInterface) *OrbitController {
	return &OrbitController{
		target:          target,
		distance:        500,
		rotation:        IdentityQuaternion(),
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
func (controller *OrbitController) Rotate(q Quaternion) {
	controller.rotation = q.Multiply(controller.rotation)
	controller.Update()
}

// SetRotation sets the absolute rotation
func (controller *OrbitController) SetRotation(q Quaternion) {
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
	qYaw := FromAxisAngle(Point3D{X: 0, Y: 1, Z: 0}, float64(dx)*yawSensitivity)
	qPitch := FromAxisAngle(Point3D{X: 1, Y: 0, Z: 0}, float64(-dy)*pitchSensitivity)
	controller.Rotate(qYaw.Multiply(qPitch))
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
	pos := Point3D{X: 0, Y: 0, Z: controller.distance}
	pos = controller.rotation.RotatePoint(pos)
	pos.Add(center)
	controller.camera.Position = pos
}

func (controller *OrbitController) lookAtTarget() {
	if controller.camera == nil || controller.target == nil {
		return
	}
	center := controller.target.GetPosition()
	cameraPos := controller.camera.Position
	dir := Point3D{
		X: center.X - cameraPos.X,
		Y: center.Y - cameraPos.Y,
		Z: center.Z - cameraPos.Z,
	}
	length := math.Sqrt(float64(dir.X*dir.X + dir.Y*dir.Y + dir.Z*dir.Z))
	if length == 0 {
		controller.camera.Rotation = IdentityQuaternion()
		return
	}
	dir.X /= Unit(length)
	dir.Y /= Unit(length)
	dir.Z /= Unit(length)

	up := Point3D{X: 0, Y: 1, Z: 0}
	if math.Abs(float64(dir.X*up.X+dir.Y*up.Y+dir.Z*up.Z)) > 0.999 {
		up = Point3D{X: 0, Y: 0, Z: 1}
	}
	side := up.Cross(dir)
	upn := dir.Cross(side)
	m00 := float64(side.X)
	m01 := float64(upn.X)
	m02 := -float64(dir.X)
	m10 := float64(side.Y)
	m11 := float64(upn.Y)
	m12 := -float64(dir.Y)
	m20 := float64(side.Z)
	m21 := float64(upn.Z)
	m22 := -float64(dir.Z)
	tr := m00 + m11 + m22
	var q Quaternion
	if tr > 0 {
		s := math.Sqrt(tr+1.0) * 2
		q.W = 0.25 * s
		q.X = (m21 - m12) / s
		q.Y = (m02 - m20) / s
		q.Z = (m10 - m01) / s
	} else if m00 > m11 && m00 > m22 {
		s := math.Sqrt(1.0+m00-m11-m22) * 2
		q.W = (m21 - m12) / s
		q.X = 0.25 * s
		q.Y = (m01 + m10) / s
		q.Z = (m02 + m20) / s
	} else if m11 > m22 {
		s := math.Sqrt(1.0+m11-m00-m22) * 2
		q.W = (m02 - m20) / s
		q.X = (m01 + m10) / s
		q.Y = 0.25 * s
		q.Z = (m12 + m21) / s
	} else {
		s := math.Sqrt(1.0+m22-m00-m11) * 2
		q.W = (m10 - m01) / s
		q.X = (m02 + m20) / s
		q.Y = (m12 + m21) / s
		q.Z = 0.25 * s
	}
	q.Normalize()
	controller.camera.Rotation = q
}

// ManualController is a controller that allows the camera to be manually controlled. Useful for debugging
type ManualController struct {
	BaseController
}

// NewManualController creates a new ManualController
func NewManualController() *ManualController {
	return &ManualController{}
}

// GetRotationSlider returns a container with sliders for controlling the rotation of the camera
func (controller *ManualController) GetRotationSlider() *fyne.Container {
	sliderYaw := widget.NewSlider(0, 360)
	sliderYaw.OnChanged = func(value float64) {
		q := FromAxisAngle(Point3D{X: 0, Y: 1, Z: 0}, value*math.Pi/180)
		controller.camera.Rotation = q.Multiply(controller.camera.Rotation)
	}
	sliderPitch := widget.NewSlider(0, 360)
	sliderPitch.OnChanged = func(value float64) {
		q := FromAxisAngle(Point3D{X: 1, Y: 0, Z: 0}, value*math.Pi/180)
		controller.camera.Rotation = q.Multiply(controller.camera.Rotation)
	}
	sliderRoll := widget.NewSlider(0, 360)
	sliderRoll.OnChanged = func(value float64) {
		q := FromAxisAngle(Point3D{X: 0, Y: 0, Z: 1}, value*math.Pi/180)
		controller.camera.Rotation = q.Multiply(controller.camera.Rotation)
	}
	sliderContainer := container.NewVBox(sliderYaw, sliderPitch, sliderRoll)
	return sliderContainer
}

// GetPositionControl returns a container with sliders for controlling the position of the camera
func (controller *ManualController) GetPositionControl() *fyne.Container {
	sliderX := widget.NewSlider(-100, 100)
	sliderX.OnChanged = func(value float64) {
		if value > 0 {
			controller.camera.Position.X += 10
		} else {
			controller.camera.Position.X -= 10
		}
	}
	sliderX.OnChangeEnded = func(value float64) {
		sliderX.Value = 0
	}

	sliderY := widget.NewSlider(-100, 100)
	sliderY.OnChanged = func(value float64) {
		if value > 0 {
			controller.camera.Position.Y += 10
		} else {
			controller.camera.Position.Y -= 10
		}
	}
	sliderY.OnChangeEnded = func(value float64) {
		sliderY.Value = 0
	}

	sliderZ := widget.NewSlider(-100, 100)
	sliderZ.OnChanged = func(value float64) {
		if value > 0 {
			controller.camera.Position.Z += 10
		} else {
			controller.camera.Position.Z -= 10
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
			label.SetText(fmt.Sprintf("X: %.2f Y: %.2f Z: %.2f      Q: (%.2f, %.2f, %.2f, %.2f)",
				controller.camera.Position.X, controller.camera.Position.Y, controller.camera.Position.Z,
				q.W, q.X, q.Y, q.Z))
			label.Refresh()
		}
	}()
	return label
}
