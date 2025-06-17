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
	controller.updateRotation()
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

func (controller *OrbitController) updateRotation() {
	if controller.camera == nil || controller.target == nil {
		return
	}
	// Always keep the target perfectly centered on the screen by looking at the target
	center := controller.target.GetPosition()
	cameraPos := controller.camera.Position
	direction := Point3D{
		X: center.X - cameraPos.X,
		Y: center.Y - cameraPos.Y,
		Z: center.Z - cameraPos.Z,
	}
	// Normalize direction
	length := math.Sqrt(float64(direction.X*direction.X + direction.Y*direction.Y + direction.Z*direction.Z))
	if length == 0 {
		controller.camera.Rotation = IdentityQuaternion()
		return
	}
	direction.X /= Unit(length)
	direction.Y /= Unit(length)
	direction.Z /= Unit(length)

	// Default up vector
	up := Point3D{X: 0, Y: 0, Z: 1}
	// If direction is parallel to up, use a different up vector
	if math.Abs(float64(direction.X*up.X+direction.Y*up.Y+direction.Z*up.Z)) > 0.999 {
		up = Point3D{X: 0, Y: 1, Z: 0}
	}

	// Calculate right and true up
	right := up.Cross(direction)
	// Normalize right
	rightLen := math.Sqrt(float64(right.X*right.X + right.Y*right.Y + right.Z*right.Z))
	if rightLen == 0 {
		right = Point3D{X: 1, Y: 0, Z: 0}
	} else {
		right.X /= Unit(rightLen)
		right.Y /= Unit(rightLen)
		right.Z /= Unit(rightLen)
	}
	trueUp := direction.Cross(right)

	// Build rotation matrix (columns: right, trueUp, direction)
	m00 := float64(right.X)
	m01 := float64(trueUp.X)
	m02 := float64(direction.X)
	m10 := float64(right.Y)
	m11 := float64(trueUp.Y)
	m12 := float64(direction.Y)
	m20 := float64(right.Z)
	m21 := float64(trueUp.Z)
	m22 := float64(direction.Z)

	// Convert rotation matrix to quaternion
	trace := m00 + m11 + m22
	var qw, qx, qy, qz float64
	if trace > 0 {
		s := math.Sqrt(trace+1.0) * 2
		qw = 0.25 * s
		qx = (m21 - m12) / s
		qy = (m02 - m20) / s
		qz = (m10 - m01) / s
	} else if m00 > m11 && m00 > m22 {
		s := math.Sqrt(1.0+m00-m11-m22) * 2
		qw = (m21 - m12) / s
		qx = 0.25 * s
		qy = (m01 + m10) / s
		qz = (m02 + m20) / s
	} else if m11 > m22 {
		s := math.Sqrt(1.0+m11-m00-m22) * 2
		qw = (m02 - m20) / s
		qx = (m01 + m10) / s
		qy = 0.25 * s
		qz = (m12 + m21) / s
	} else {
		s := math.Sqrt(1.0+m22-m00-m11) * 2
		qw = (m10 - m01) / s
		qx = (m02 + m20) / s
		qy = (m12 + m21) / s
		qz = 0.25 * s
	}
	controller.camera.Rotation = NewQuaternion(qw, qx, qy, qz)
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
