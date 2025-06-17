package types

import (
	"math"
)

type Vec3 struct {
	X, Y, Z float64
}

type Mat4 [4][4]float64

func (v Vec3) Subtract(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Vec3) Normalize() Vec3 {
	length := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	return Vec3{v.X / length, v.Y / length, v.Z / length}
}

func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		v.Y*other.Z - v.Z*other.Y,
		v.Z*other.X - v.X*other.Z,
		v.X*other.Y - v.Y*other.X,
	}
}

func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func LookAt(eyePoint3D Point3D, centerPoint3D Point3D, upDirectionVector DirectionVector) Mat4 {
	eye := Vec3{float64(eyePoint3D.X), float64(eyePoint3D.Y), float64(eyePoint3D.Z)}
	center := Vec3{float64(centerPoint3D.X), float64(centerPoint3D.Y), float64(centerPoint3D.Z)}
	up := Vec3{float64(upDirectionVector.X), float64(upDirectionVector.Y), float64(upDirectionVector.Z)}

	f := center.Subtract(eye).Normalize()
	u := up.Normalize()
	s := f.Cross(u).Normalize()
	u = s.Cross(f)

	return Mat4{
		{s.X, u.X, -f.X, 0},
		{s.Y, u.Y, -f.Y, 0},
		{s.Z, u.Z, -f.Z, 0},
		{-s.Dot(eye), -u.Dot(eye), f.Dot(eye), 1},
	}
}

type Vec4 struct {
	X, Y, Z, W float64
}

func (m Mat4) MultiplyVec4(v Vec4) Vec4 {
	return Vec4{
		m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3]*v.W,
		m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3]*v.W,
		m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3]*v.W,
		m[3][0]*v.X + m[3][1]*v.Y + m[3][2]*v.Z + m[3][3]*v.W,
	}
}

func (m Mat4) MultiplyDirectionVector(direction DirectionVector) DirectionVector {
	v := Vec4{float64(direction.X), float64(direction.Y), float64(direction.Z), 0}

	result := m.MultiplyVec4(v)
	return DirectionVector{Point3D{Unit(result.X), Unit(result.Y), Unit(result.Z)}}
}
