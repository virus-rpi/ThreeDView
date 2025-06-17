package object

import (
	. "ThreeDView/types"
	"image/color"
	"math"
)

func NewCube(size Unit, position Point3D, rotation Quaternion, color color.Color, w ThreeDWidgetInterface) *Object {
	half := size / 2
	vertices := []Point3D{
		{X: -half, Y: -half, Z: -half},
		{X: half, Y: -half, Z: -half},
		{X: half, Y: half, Z: -half},
		{X: -half, Y: half, Z: -half},
		{X: -half, Y: -half, Z: half},
		{X: half, Y: -half, Z: half},
		{X: half, Y: half, Z: half},
		{X: -half, Y: half, Z: half},
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

		facesData[i] = FaceData{Face: [3]Point3D{p1, p2, p3}, Color: color}
	}

	cube := Object{
		Faces:    facesData,
		Position: position,
		Rotation: rotation,
		Widget:   w,
	}
	w.AddObject(&cube)
	return &cube
}

func NewPlane(size Unit, position Point3D, rotation Quaternion, color color.Color, w ThreeDWidgetInterface, resolution int) *Object {
	half := size / 2
	step := size / Unit(resolution)
	var vertices []Point3D
	for i := 0; i <= resolution; i++ {
		for j := 0; j <= resolution; j++ {
			vertices = append(vertices, Point3D{
				X: -half + Unit(i)*step,
				Y: -half + Unit(j)*step,
				Z: 0,
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

		facesData[i] = FaceData{Face: [3]Point3D{p1, p2, p3}, Color: color}
	}

	plane := Object{
		Faces:    facesData,
		Position: position,
		Rotation: rotation,
		Widget:   w,
	}
	w.AddObject(&plane)
	return &plane
}

func NewOrientationObject(w ThreeDWidgetInterface) *Object {
	size := Unit(2)
	thickness := size / 20

	faces := []FaceData{
		{
			Face: [3]Point3D{
				{X: 0, Y: -thickness, Z: -thickness},
				{X: size, Y: -thickness, Z: -thickness},
				{X: 0, Y: thickness, Z: -thickness},
			},
			Color: color.RGBA{R: 255, A: 255},
		},
		{
			Face: [3]Point3D{
				{X: size, Y: -thickness, Z: -thickness},
				{X: size, Y: thickness, Z: -thickness},
				{X: 0, Y: thickness, Z: -thickness},
			},
			Color: color.RGBA{R: 255, A: 255},
		},
		{
			Face: [3]Point3D{
				{X: -thickness, Y: 0, Z: -thickness},
				{X: -thickness, Y: size, Z: -thickness},
				{X: thickness, Y: 0, Z: -thickness},
			},
			Color: color.RGBA{R: 255, G: 255, A: 255},
		},
		{
			Face: [3]Point3D{
				{X: -thickness, Y: size, Z: -thickness},
				{X: thickness, Y: size, Z: -thickness},
				{X: thickness, Y: 0, Z: -thickness},
			},
			Color: color.RGBA{R: 255, G: 255, A: 255},
		},
		{
			Face: [3]Point3D{
				{X: -thickness, Y: -thickness, Z: 0},
				{X: -thickness, Y: -thickness, Z: size},
				{X: thickness, Y: -thickness, Z: 0},
			},
			Color: color.RGBA{B: 255, A: 255},
		},
		{
			Face: [3]Point3D{
				{X: -thickness, Y: -thickness, Z: size},
				{X: thickness, Y: -thickness, Z: size},
				{X: thickness, Y: -thickness, Z: 0},
			},
			Color: color.RGBA{B: 255, A: 255},
		},
	}

	orientationObject := Object{
		Faces:    faces,
		Position: Point3D{},
		Rotation: IdentityQuaternion(),
		Widget:   w,
	}
	w.RegisterTickMethod(func() {
		orientationObject.Position = w.GetCamera().UnProject(Point2D{X: 60, Y: 70}, 20, w.GetWidth(), w.GetHeight())
	})
	w.AddObject(&orientationObject)
	return &orientationObject
}

func NewEmpty(w ThreeDWidgetInterface, position Point3D) *Object {
	empty := Object{
		Faces:    nil,
		Position: position,
		Rotation: IdentityQuaternion(),
		Widget:   w,
	}
	w.AddObject(&empty)
	return &empty
}

func NewCylinder(position Point3D, rotation Quaternion, color color.Color, w ThreeDWidgetInterface, height, radius Unit) *Object {
	var faces []FaceData
	for i := 0; i < 360; i += 20 {
		angle1 := Degrees(i).ToRadians()
		angle2 := Degrees(i + 20).ToRadians()
		p1 := Point3D{
			X: radius * Unit(math.Cos(float64(angle1))),
			Y: radius * Unit(math.Sin(float64(angle1))),
			Z: -height / 2,
		}
		p2 := Point3D{
			X: radius * Unit(math.Cos(float64(angle1))),
			Y: radius * Unit(math.Sin(float64(angle1))),
			Z: height / 2,
		}
		p3 := Point3D{
			X: radius * Unit(math.Cos(float64(angle2))),
			Y: radius * Unit(math.Sin(float64(angle2))),
			Z: height / 2,
		}
		p4 := Point3D{
			X: radius * Unit(math.Cos(float64(angle2))),
			Y: radius * Unit(math.Sin(float64(angle2))),
			Z: -height / 2,
		}
		sideFaces := []FaceData{
			{
				Face:  [3]Point3D{p1, p2, p3},
				Color: color,
			},
			{
				Face:  [3]Point3D{p1, p3, p4},
				Color: color,
			},
		}
		faces = append(faces, sideFaces...)
	}

	cylinder := Object{
		Faces:    faces,
		Position: position,
		Rotation: rotation,
		Widget:   w,
	}
	w.AddObject(&cylinder)
	return &cylinder
}

func NewCone(position Point3D, rotation Quaternion, color color.Color, w ThreeDWidgetInterface, height, radius Unit) *Object {
	var faces []FaceData

	for i := 0; i < 360; i += 20 {
		angle1 := Degrees(i).ToRadians()
		angle2 := Degrees(i + 20).ToRadians()
		p1 := Point3D{
			X: 0,
			Y: 0,
			Z: height / 2,
		}
		p2 := Point3D{
			X: radius * Unit(math.Cos(float64(angle1))),
			Y: radius * Unit(math.Sin(float64(angle1))),
			Z: -height / 2,
		}
		p3 := Point3D{
			X: radius * Unit(math.Cos(float64(angle2))),
			Y: radius * Unit(math.Sin(float64(angle2))),
			Z: -height / 2,
		}
		sideFaces := []FaceData{
			{
				Face:  [3]Point3D{p1, p2, p3},
				Color: color,
			},
		}
		faces = append(faces, sideFaces...)
	}

	cone := Object{
		Faces:    faces,
		Position: position,
		Rotation: rotation,
		Widget:   w,
	}
	w.AddObject(&cone)
	return &cone
}
