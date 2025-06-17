package types

import (
	"math"
)

type Point3D struct {
	X, Y, Z Unit
}

// Rotate rotates the point around a pivot point by the given quaternion
func (point *Point3D) Rotate(pivot Point3D, rotation Quaternion) {
	p := *point
	p.Subtract(pivot)
	p = rotation.RotatePoint(p)
	p.Add(pivot)
	*point = p
}

// Add adds another point to the point
func (point *Point3D) Add(other Point3D) {
	point.X += other.X
	point.Y += other.Y
	point.Z += other.Z
}

// Subtract subtracts another point from the point
func (point *Point3D) Subtract(other Point3D) {
	point.X -= other.X
	point.Y -= other.Y
	point.Z -= other.Z
}

// DistanceTo returns the distance between the point and another point
func (point *Point3D) DistanceTo(other Point3D) Unit {
	dx := point.X - other.X
	dy := point.Y - other.Y
	dz := point.Z - other.Z
	return Unit(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Dot returns the dot product of the point with another point
func (point *Point3D) Dot(other Point3D) Unit {
	return point.X*other.X + point.Y*other.Y + point.Z*other.Z
}

// Cross returns the cross-product of the point with another point
func (point *Point3D) Cross(other Point3D) Point3D {
	return Point3D{
		X: point.Y*other.Z - point.Z*other.Y,
		Y: point.Z*other.X - point.X*other.Z,
		Z: point.X*other.Y - point.Y*other.X,
	}
}
