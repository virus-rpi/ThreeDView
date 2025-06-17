package types

import (
	"math"
)

// Quaternion represents a quaternion for 3D rotations
// w + xi + yj + zk
// All fields are float64 for precision
// Use Unit for conversion to/from Point3D if needed

type Quaternion struct {
	W, X, Y, Z float64
}

// NewQuaternion creates a new quaternion from w, x, y, z
func NewQuaternion(w, x, y, z float64) Quaternion {
	return Quaternion{W: w, X: x, Y: y, Z: z}
}

// IdentityQuaternion returns the identity quaternion (no rotation)
func IdentityQuaternion() Quaternion {
	return Quaternion{W: 1, X: 0, Y: 0, Z: 0}
}

// FromAxisAngle creates a quaternion from an axis (normalized) and angle (in radians)
func FromAxisAngle(axis Point3D, angle float64) Quaternion {
	half := angle / 2
	sinHalf := math.Sin(half)
	return Quaternion{
		W: math.Cos(half),
		X: float64(axis.X) * sinHalf,
		Y: float64(axis.Y) * sinHalf,
		Z: float64(axis.Z) * sinHalf,
	}
}

// Normalize normalizes the quaternion
func (q *Quaternion) Normalize() {
	mag := math.Sqrt(q.W*q.W + q.X*q.X + q.Y*q.Y + q.Z*q.Z)
	if mag == 0 {
		q.W = 1
		q.X = 0
		q.Y = 0
		q.Z = 0
		return
	}
	q.W /= mag
	q.X /= mag
	q.Y /= mag
	q.Z /= mag
}

// Multiply multiplies two quaternions (q * r)
func (q Quaternion) Multiply(r Quaternion) Quaternion {
	return Quaternion{
		W: q.W*r.W - q.X*r.X - q.Y*r.Y - q.Z*r.Z,
		X: q.W*r.X + q.X*r.W + q.Y*r.Z - q.Z*r.Y,
		Y: q.W*r.Y - q.X*r.Z + q.Y*r.W + q.Z*r.X,
		Z: q.W*r.Z + q.X*r.Y - q.Y*r.X + q.Z*r.W,
	}
}

// Conjugate returns the conjugate of the quaternion
func (q Quaternion) Conjugate() Quaternion {
	return Quaternion{W: q.W, X: -q.X, Y: -q.Y, Z: -q.Z}
}

// RotatePoint rotates a Point3D by this quaternion (assumes normalized)
func (q Quaternion) RotatePoint(p Point3D) Point3D {
	qx := q.X
	qy := q.Y
	qz := q.Z
	qw := q.W
	x := float64(p.X)
	y := float64(p.Y)
	z := float64(p.Z)

	// Quaternion multiplication: q * v * q^-1
	// v as quaternion: (0, x, y, z)
	// q^-1 = conjugate if normalized
	//
	// result = q * v * q^-1
	//
	ix := qw*x + qy*z - qz*y
	iy := qw*y + qz*x - qx*z
	iz := qw*z + qx*y - qy*x
	iw := -qx*x - qy*y - qz*z

	return Point3D{
		X: Unit(ix*qw + iw*-qx + iy*-qz - iz*-qy),
		Y: Unit(iy*qw + iw*-qy + iz*-qx - ix*-qz),
		Z: Unit(iz*qw + iw*-qz + ix*-qy - iy*-qx),
	}
}

// ToRotationMatrix converts the quaternion to a 3x3 rotation matrix
func (q Quaternion) ToRotationMatrix() RotationMatrix {
	qx := q.X
	qy := q.Y
	qz := q.Z
	qw := q.W
	return RotationMatrix{
		{1 - 2*qy*qy - 2*qz*qz, 2*qx*qy - 2*qz*qw, 2*qx*qz + 2*qy*qw},
		{2*qx*qy + 2*qz*qw, 1 - 2*qx*qx - 2*qz*qz, 2*qy*qz - 2*qx*qw},
		{2*qx*qz - 2*qy*qw, 2*qy*qz + 2*qx*qw, 1 - 2*qx*qx - 2*qy*qy},
	}
}

// FromEuler creates a quaternion from Euler angles (in radians)
func FromEuler(yaw, pitch, roll float64) Quaternion {
	cy := math.Cos(yaw * 0.5)
	sy := math.Sin(yaw * 0.5)
	cp := math.Cos(pitch * 0.5)
	sp := math.Sin(pitch * 0.5)
	cr := math.Cos(roll * 0.5)
	sr := math.Sin(roll * 0.5)
	return Quaternion{
		W: cr*cp*cy + sr*sp*sy,
		X: sr*cp*cy - cr*sp*sy,
		Y: cr*sp*cy + sr*cp*sy,
		Z: cr*cp*sy - sr*sp*cy,
	}
}

// ToEuler returns yaw, pitch, roll (in radians)
func (q Quaternion) ToEuler() (yaw, pitch, roll float64) {
	// yaw (Z), pitch (Y), roll (X)
	sinr_cosp := 2 * (q.W*q.X + q.Y*q.Z)
	cosr_cosp := 1 - 2*(q.X*q.X+q.Y*q.Y)
	roll = math.Atan2(sinr_cosp, cosr_cosp)

	sinp := 2 * (q.W*q.Y - q.Z*q.X)
	if math.Abs(sinp) >= 1 {
		pitch = math.Copysign(math.Pi/2, sinp)
	} else {
		pitch = math.Asin(sinp)
	}

	siny_cosp := 2 * (q.W*q.Z + q.X*q.Y)
	cosy_cosp := 1 - 2*(q.Y*q.Y+q.Z*q.Z)
	yaw = math.Atan2(siny_cosp, cosy_cosp)
	return
}
