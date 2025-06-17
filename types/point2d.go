package types

// Point2D represents a point in 2D space
type Point2D struct {
	X, Y Pixel
}

// InBounds returns true if the point is in bounds of the given width and height
func (point *Point2D) InBounds(x1, y1, x2, y2 Pixel) bool {
	return point.X >= x1 && point.X < x2 && point.Y >= y1 && point.Y < y2
}

// Face represents a face in 3D space as 3 3D points
type Face [3]Point3D

// Rotate rotates the face around a pivot point by the given quaternion rotation
func (face *Face) Rotate(pivot Point3D, rotation Quaternion) {
	for i := 0; i < 3; i++ {
		p := face[i]
		p.Subtract(pivot)
		p = rotation.RotatePoint(p)
		p.Add(pivot)
		face[i] = p
	}
}

// Add adds another point to the face
func (face *Face) Add(other Point3D) {
	face[0].Add(other)
	face[1].Add(other)
	face[2].Add(other)
}

// DistanceTo returns the distance between the face and a point
func (face *Face) DistanceTo(point Point3D) Unit {
	return (face[0].DistanceTo(point) + face[1].DistanceTo(point) + face[2].DistanceTo(point)) / 3
}
