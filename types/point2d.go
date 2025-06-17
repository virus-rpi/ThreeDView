package types

// Point2D represents a point in 2D space
type Point2D struct {
	X, Y Pixel
}

// InBounds returns true if the point is in bounds of the given width and height
func (point *Point2D) InBounds(x1, y1, x2, y2 Pixel) bool {
	return point.X >= x1 && point.X < x2 && point.Y >= y1 && point.Y < y2
}
