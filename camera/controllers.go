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
	controller.Move(Unit(-y * 5)) // Negative to zoom in on scroll up
}

// Update recalculates the camera's position and orientation
func (controller *OrbitController) Update() {
	controller.updatePosition()
}

func (controller *OrbitController) updatePosition() {
	if controller.camera == nil || controller.target == nil {
		return
	}
	center := controller.target.GetPosition()
	// Camera offset: start at (0, 0, distance) and rotate by current quaternion
	pos := Point3D{X: 0, Y: 0, Z: controller.distance}
	pos = controller.rotation.RotatePoint(pos)
	pos.Add(center)
	controller.camera.Position = pos

	// Camera should look at the target: compute look rotation quaternion
	lookDir := Point3D{X: center.X - pos.X, Y: center.Y - pos.Y, Z: center.Z - pos.Z}
	up := Point3D{X: 0, Y: 1, Z: 0}
	controller.camera.Rotation = LookRotationQuaternion(lookDir, up)
}

// LookRotationQuaternion returns a quaternion that rotates the forward vector (0,0,1) to lookDir, using up as the up direction
func LookRotationQuaternion(lookDir, up Point3D) Quaternion {
	// Normalize lookDir
	mag := math.Sqrt(float64(lookDir.X*lookDir.X + lookDir.Y*lookDir.Y + lookDir.Z*lookDir.Z))
	if mag == 0 {
		return IdentityQuaternion()
	}
	f := Point3D{X: lookDir.X / Unit(mag), Y: lookDir.Y / Unit(mag), Z: lookDir.Z / Unit(mag)}
	// Normalize up
	upMag := math.Sqrt(float64(up.X*up.X + up.Y*up.Y + up.Z*up.Z))
	if upMag == 0 {
		up = Point3D{X: 0, Y: 1, Z: 0}
	} else {
		up = Point3D{X: up.X / Unit(upMag), Y: up.Y / Unit(upMag), Z: up.Z / Unit(upMag)}
	}
	// Right vector
	r := f.Cross(up)
	// Recompute up to ensure orthogonality
	u := r.Cross(f)
	// Rotation matrix (columns: right, up, -forward)
	m := [3][3]float64{
		{float64(r.X), float64(u.X), -float64(f.X)},
		{float64(r.Y), float64(u.Y), -float64(f.Y)},
		{float64(r.Z), float64(u.Z), -float64(f.Z)},
	}
	// Convert rotation matrix to quaternion
	tr := m[0][0] + m[1][1] + m[2][2]
	var q Quaternion
	if tr > 0 {
		s := math.Sqrt(tr+1.0) * 2
		q.W = 0.25 * s
		q.X = (m[2][1] - m[1][2]) / s
		q.Y = (m[0][2] - m[2][0]) / s
		q.Z = (m[1][0] - m[0][1]) / s
	} else if m[0][0] > m[1][1] && m[0][0] > m[2][2] {
		s := math.Sqrt(1.0+m[0][0]-m[1][1]-m[2][2]) * 2
		q.W = (m[2][1] - m[1][2]) / s
		q.X = 0.25 * s
		q.Y = (m[0][1] + m[1][0]) / s
		q.Z = (m[0][2] + m[2][0]) / s
	} else if m[1][1] > m[2][2] {
		s := math.Sqrt(1.0+m[1][1]-m[0][0]-m[2][2]) * 2
		q.W = (m[0][2] - m[2][0]) / s
		q.X = (m[0][1] + m[1][0]) / s
		q.Y = 0.25 * s
		q.Z = (m[1][2] + m[2][1]) / s
	} else {
		s := math.Sqrt(1.0+m[2][2]-m[0][0]-m[1][1]) * 2
		q.W = (m[1][0] - m[0][1]) / s
		q.X = (m[0][2] + m[2][0]) / s
		q.Y = (m[1][2] + m[2][1]) / s
		q.Z = 0.25 * s
	}
	return q
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
