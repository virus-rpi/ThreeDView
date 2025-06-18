package types

import mgl "github.com/go-gl/mathgl/mgl64"

// Face represents a face in 3D space as 3 3D points
type Face [3]mgl.Vec3

// Rotate rotates the face around a pivot point by the given quaternion rotation
func (face *Face) Rotate(pivot mgl.Vec3, rotation mgl.Quat) {
	for i := 0; i < len(face); i++ {
		p := face[i].Sub(pivot)
		p = rotation.Rotate(p)
		face[i] = p.Add(pivot)
	}
}

// Add adds another point to the face
func (face *Face) Add(other mgl.Vec3) {
	for i := 0; i < len(face); i++ {
		face[i] = face[i].Add(other)
	}
}

// DistanceTo returns the distance between the face and a point
func (face *Face) DistanceTo(point mgl.Vec3) Unit {
	totalDistance := Unit(0)
	for _, vertex := range face {
		totalDistance += Unit(vertex.Sub(point).Len())
	}
	return totalDistance / Unit(len(face))
}

// Normal calculates the normal vector of the face
func (face *Face) Normal() mgl.Vec3 {
	edge1 := face[1].Sub(face[0])
	edge2 := face[2].Sub(face[0])
	normal := edge1.Cross(edge2).Normalize()
	return normal
}
