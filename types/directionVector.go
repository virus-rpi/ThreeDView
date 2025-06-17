package types

import "math"

// DirectionVector represents a vector in 3D space as a normalized vector
type DirectionVector struct {
	Point3D
}

// Magnitude returns the magnitude of the point (distance from origin)
func (vector *DirectionVector) Magnitude() Unit {
	return Unit(math.Sqrt(float64(vector.X*vector.X + vector.Y*vector.Y + vector.Z*vector.Z)))
}

// Normalize normalizes the point
func (vector *DirectionVector) Normalize() {
	magnitude := vector.Magnitude()
	if magnitude == 0 {
		return
	}
	vector.X /= magnitude
	vector.Y /= magnitude
	vector.Z /= magnitude
}

// ToQuaternion returns a quaternion that rotates the -Z axis to this direction vector
func (vector *DirectionVector) ToQuaternion() Quaternion {
	forward := Point3D{X: 0, Y: 0, Z: -1}
	d := Point3D{X: vector.X, Y: vector.Y, Z: vector.Z}
	if math.Abs(float64(d.X-forward.X))+math.Abs(float64(d.Y-forward.Y))+math.Abs(float64(d.Z-forward.Z)) < 1e-6 {
		return IdentityQuaternion()
	}
	if math.Abs(float64(d.X+forward.X))+math.Abs(float64(d.Y+forward.Y))+math.Abs(float64(d.Z+forward.Z)) < 1e-6 {
		return FromAxisAngle(Point3D{X: 0, Y: 1, Z: 0}, math.Pi)
	}
	axis := DirectionVector{forward.Cross(d)}
	axis.Normalize()
	angle := math.Acos(float64(forward.Dot(d)))
	return FromAxisAngle(axis.Point3D, angle)
}
