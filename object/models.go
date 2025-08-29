package object

import (
	mgl "github.com/go-gl/mathgl/mgl64"
	. "github.com/virus-rpi/ThreeDView/types"
	"image/color"
	"math"
)

// NewCube creates a cube object with the given size, position, rotation, and color
func NewCube(size Unit, position mgl.Vec3, rotation mgl.Quat, color color.Color, w ThreeDWidgetInterface) *Object {
	half := float64(size) / 2
	vertices := []mgl.Vec3{
		{-half, -half, -half},
		{half, -half, -half},
		{half, half, -half},
		{-half, half, -half},
		{-half, -half, half},
		{half, -half, half},
		{half, half, half},
		{-half, half, half},
	}
	faces := [][3]int{
		{0, 1, 2}, {0, 2, 3},
		{4, 5, 6}, {4, 6, 7},
		{0, 1, 5}, {0, 5, 4},
		{2, 3, 7}, {2, 7, 6},
		{0, 3, 7}, {0, 7, 4},
		{1, 2, 6}, {1, 6, 5},
	}

	var facesData = make([]FaceData, len(faces))
	for i := 0; i < len(faces); i++ {
		face := faces[i]
		p1 := vertices[face[0]]
		p2 := vertices[face[1]]
		p3 := vertices[face[2]]

		facesData[i] = FaceData{
			Face:  [3]mgl.Vec3{p1, p2, p3},
			Color: color,
		}
	}

	cube := Object{
		faces:    facesData,
		position: position,
		rotation: rotation,
		widget:   w,
	}
	w.AddObject(&cube)
	return &cube
}

// NewPlane creates a plane object with the given size, position, rotation, color, and resolution
func NewPlane(size Unit, position mgl.Vec3, rotation mgl.Quat, color color.Color, w ThreeDWidgetInterface, resolution int) *Object {
	half := size / 2
	step := size / Unit(resolution)
	var vertices []mgl.Vec3
	for i := 0; i <= resolution; i++ {
		for j := 0; j <= resolution; j++ {
			vertices = append(vertices, mgl.Vec3{
				float64(-half + Unit(i)*step),
				0,
				float64(-half + Unit(j)*step),
			})
		}
	}

	var faces [][3]int
	for i := 0; i < resolution; i++ {
		for j := 0; j < resolution; j++ {
			topLeft := i*(resolution+1) + j
			topRight := topLeft + 1
			bottomLeft := topLeft + (resolution + 1)
			bottomRight := bottomLeft + 1

			faces = append(faces, [3]int{topLeft, topRight, bottomRight})
			faces = append(faces, [3]int{topLeft, bottomRight, bottomLeft})
		}
	}

	var facesData = make([]FaceData, len(faces))
	for i := 0; i < len(faces); i++ {
		face := faces[i]
		p1 := vertices[face[0]]
		p2 := vertices[face[1]]
		p3 := vertices[face[2]]

		facesData[i] = FaceData{Face: [3]mgl.Vec3{p1, p2, p3}, Color: color}
	}

	plane := Object{
		faces:    facesData,
		position: position,
		rotation: rotation,
		widget:   w,
	}
	w.AddObject(&plane)
	return &plane
}

// NewOrientationObject creates an orientation object with arrows for X, Y, and Z axes
func NewOrientationObject(w ThreeDWidgetInterface) *Object {
	size := Unit(5)
	thickness := size / 20

	faces := []FaceData{
		{
			Face: [3]mgl.Vec3{
				{0, -float64(thickness), -float64(thickness)},
				{float64(size), -float64(thickness), -float64(thickness)},
				{0, float64(thickness), -float64(thickness)},
			},
			Color: color.RGBA{R: 255, A: 255},
		},
		{
			Face: [3]mgl.Vec3{
				{float64(size), -float64(thickness), -float64(thickness)},
				{float64(size), float64(thickness), -float64(thickness)},
				{0, float64(thickness), -float64(thickness)},
			},
			Color: color.RGBA{R: 255, A: 255},
		},
		{
			Face: [3]mgl.Vec3{
				{-float64(thickness), 0, -float64(thickness)},
				{-float64(thickness), float64(size), -float64(thickness)},
				{float64(thickness), 0, -float64(thickness)},
			},
			Color: color.RGBA{R: 255, G: 255, A: 255},
		},
		{
			Face: [3]mgl.Vec3{
				{-float64(thickness), float64(size), -float64(thickness)},
				{float64(thickness), float64(size), -float64(thickness)},
				{float64(thickness), 0, -float64(thickness)},
			},
			Color: color.RGBA{R: 255, G: 255, A: 255},
		},
		{
			Face: [3]mgl.Vec3{
				{-float64(thickness), -float64(thickness), 0},
				{-float64(thickness), -float64(thickness), float64(size)},
				{float64(thickness), -float64(thickness), 0},
			},
			Color: color.RGBA{B: 255, A: 255},
		},
		{
			Face: [3]mgl.Vec3{
				{-float64(thickness), -float64(thickness), float64(size)},
				{float64(thickness), -float64(thickness), float64(size)},
				{float64(thickness), -float64(thickness), 0},
			},
			Color: color.RGBA{B: 255, A: 255},
		},
	}

	orientationObject := Object{
		faces:    faces,
		position: mgl.Vec3{0, 0, 0},
		rotation: mgl.QuatIdent(),
		widget:   w,
	}
	w.RegisterTickMethod(func() {
		desiredPixelSize := 40
		screenHeight := w.GetHeight()
		fov := float64(w.GetCamera().Fov())

		fovRad := fov * math.Pi / 180.0
		distance := (float64(size) / 2) / (math.Tan(fovRad/2) * (float64(desiredPixelSize) / float64(screenHeight)))

		margin := desiredPixelSize * 2
		orientationObject.SetPosition(w.GetCamera().UnProject(mgl.Vec2{float64(margin), float64(margin)}, Unit(distance)))
	})
	w.AddObject(&orientationObject)
	return &orientationObject
}

// NewEmpty creates an empty Object with no faces, useful for markers or placeholders
func NewEmpty(w ThreeDWidgetInterface, position mgl.Vec3) *Object {
	empty := Object{
		faces:    nil,
		position: position,
		rotation: mgl.QuatIdent(),
		widget:   w,
	}
	w.AddObject(&empty)
	return &empty
}

// NewCylinder creates a sphere object with the given position, rotation, color, and radius
func NewCylinder(position mgl.Vec3, rotation mgl.Quat, color color.Color, w ThreeDWidgetInterface, height, radius Unit) *Object {
	var faces []FaceData
	for i := 0; i < 360; i += 20 {
		angle1 := float64(i) * math.Pi / 180
		angle2 := float64(i+20) * math.Pi / 180
		p1 := mgl.Vec3{
			float64(radius) * math.Cos(angle1),
			float64(radius) * math.Sin(angle1),
			-float64(height) / 2,
		}
		p2 := mgl.Vec3{
			float64(radius) * math.Cos(angle1),
			float64(radius) * math.Sin(angle1),
			float64(height) / 2,
		}
		p3 := mgl.Vec3{
			float64(radius) * math.Cos(angle2),
			float64(radius) * math.Sin(angle2),
			float64(height) / 2,
		}
		p4 := mgl.Vec3{
			float64(radius) * math.Cos(angle2),
			float64(radius) * math.Sin(angle2),
			-float64(height) / 2,
		}
		sideFaces := []FaceData{
			{
				Face:  [3]mgl.Vec3{p1, p2, p3},
				Color: color,
			},
			{
				Face:  [3]mgl.Vec3{p1, p3, p4},
				Color: color,
			},
		}
		faces = append(faces, sideFaces...)
	}

	cylinder := Object{
		faces:    faces,
		position: position,
		rotation: rotation,
		widget:   w,
	}
	w.AddObject(&cylinder)
	return &cylinder
}

// NewCone creates a cone object with the given position, rotation, color, and radius
func NewCone(position mgl.Vec3, rotation mgl.Quat, color color.Color, w ThreeDWidgetInterface, height, radius Unit) *Object {
	var faces []FaceData

	for i := 0; i < 360; i += 20 {
		angle1 := float64(i) * math.Pi / 180
		angle2 := float64(i+20) * math.Pi / 180
		p1 := mgl.Vec3{
			0,
			0,
			float64(height) / 2,
		}
		p2 := mgl.Vec3{
			float64(radius) * math.Cos(angle1),
			float64(radius) * math.Sin(angle1),
			-float64(height) / 2,
		}
		p3 := mgl.Vec3{
			float64(radius) * math.Cos(angle2),
			float64(radius) * math.Sin(angle2),
			-float64(height) / 2,
		}
		sideFaces := []FaceData{
			{
				Face:  [3]mgl.Vec3{p1, p2, p3},
				Color: color,
			},
		}
		faces = append(faces, sideFaces...)
	}

	cone := Object{
		faces:    faces,
		position: position,
		rotation: rotation,
		widget:   w,
	}
	w.AddObject(&cone)
	return &cone
}
