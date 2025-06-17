package types

import "math"

// Unit is the unit for distance in 3D space
type Unit float64

// Pixel is the unit for distance in 2D space
type Pixel int

// Degrees represents an angle in degrees
type Degrees int

// ToRadians converts degrees to radians
func (degrees Degrees) ToRadians() Radians {
	return Radians(float64(degrees) * math.Pi / 180)
}

// Radians represents an angle in radians
type Radians float64

// ToDegrees converts radians to degrees
func (radians Radians) ToDegrees() Degrees {
	return Degrees(radians * 180 / math.Pi)
}
