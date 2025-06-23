package types

import mgl "github.com/go-gl/mathgl/mgl64"

type AABB struct {
	Min mgl.Vec3
	Max mgl.Vec3
}

// Contains checks if an AABB is fully contained within another AABB
func (a *AABB) Contains(b AABB) bool {
	return a.Min.X() <= b.Min.X() && a.Max.X() >= b.Max.X() &&
		a.Min.Y() <= b.Min.Y() && a.Max.Y() >= b.Max.Y() &&
		a.Min.Z() <= b.Min.Z() && a.Max.Z() >= b.Max.Z()
}

// Center returns the center of the AABB
func (a *AABB) Center() mgl.Vec3 {
	return a.Min.Add(a.Max).Mul(0.5)
}

// Size returns the size of the AABB
func (a *AABB) Size() mgl.Vec3 {
	return a.Max.Sub(a.Min)
}
