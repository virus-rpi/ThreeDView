package types

import "math"

// RotationMatrix represents a 3x3 rotation matrix
type RotationMatrix [3][3]float64

// ApplyInverseRotationMatrix applies the inverse of the rotation matrix to a point
func (rotationMatrix *RotationMatrix) ApplyInverseRotationMatrix(point Point3D) Point3D {
	return Point3D{
		X: Unit(rotationMatrix[0][0]*float64(point.X) + rotationMatrix[0][1]*float64(point.Y) + rotationMatrix[0][2]*float64(point.Z)),
		Y: Unit(rotationMatrix[1][0]*float64(point.X) + rotationMatrix[1][1]*float64(point.Y) + rotationMatrix[1][2]*float64(point.Z)),
		Z: Unit(rotationMatrix[2][0]*float64(point.X) + rotationMatrix[2][1]*float64(point.Y) + rotationMatrix[2][2]*float64(point.Z)),
	}
}

// Transpose transposes the rotation matrix
func (rotationMatrix *RotationMatrix) Transpose() RotationMatrix {
	return RotationMatrix{
		{rotationMatrix[0][0], rotationMatrix[1][0], rotationMatrix[2][0]},
		{rotationMatrix[0][1], rotationMatrix[1][1], rotationMatrix[2][1]},
		{rotationMatrix[0][2], rotationMatrix[1][2], rotationMatrix[2][2]},
	}
}

// Multiply multiplies the rotation matrix with another rotation matrix
func (rotationMatrix *RotationMatrix) Multiply(other RotationMatrix) RotationMatrix {
	result := RotationMatrix{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				result[i][j] += rotationMatrix[i][k] * other[k][j]
			}
		}
	}
	return result
}

// ToRotation3D converts the rotation matrix to a Rotation3D
func (rotationMatrix *RotationMatrix) ToRotation3D() Rotation3D {
	rotation := Rotation3D{}
	rotation.Roll = Degrees(math.Asin(rotationMatrix[1][2]))
	rotation.Pitch = Degrees(math.Atan2(rotationMatrix[0][2], rotationMatrix[2][2]))
	rotation.Yaw = Degrees(math.Atan2(rotationMatrix[1][0], rotationMatrix[1][1]))
	return rotation
}
