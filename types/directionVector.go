package types

import "math"

// DirectionVector represents a vector in 3D space as a normalized vector
type DirectionVector struct {
	Point3D
}

// ToRotation converts a DirectionVector to a rotation in degree
func (point *DirectionVector) ToRotation() Rotation3D {
	point.Normalize()
	return Rotation3D{
		Yaw:   Radians(math.Asin(float64(point.Y))).ToDegrees(),
		Pitch: Radians(math.Atan2(float64(point.X), float64(point.Z))).ToDegrees(),
	}
}

// Magnitude returns the magnitude of the point (distance from origin)
func (point *DirectionVector) Magnitude() Unit {
	return Unit(math.Sqrt(float64(point.X*point.X + point.Y*point.Y + point.Z*point.Z)))
}

// Normalize normalizes the point
func (point *DirectionVector) Normalize() {
	magnitude := point.Magnitude()
	if magnitude == 0 {
		return
	}
	point.X /= magnitude
	point.Y /= magnitude
	point.Z /= magnitude
}
