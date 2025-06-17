package types

// Face represents a face in 3D space as 3 3D points
type Face [3]Point3D

// Rotate rotates the face around a pivot point by the given quaternion rotation
func (face *Face) Rotate(pivot Point3D, rotation Quaternion) {
	for i := 0; i < len(face); i++ {
		face[i].Rotate(pivot, rotation)
	}
}

// Add adds another point to the face
func (face *Face) Add(other Point3D) {
	for i := 0; i < len(face); i++ {
		face[i].Add(other)
	}
}

// DistanceTo returns the distance between the face and a point
func (face *Face) DistanceTo(point Point3D) Unit {
	return (face[0].DistanceTo(point) + face[1].DistanceTo(point) + face[2].DistanceTo(point)) / 3
}
