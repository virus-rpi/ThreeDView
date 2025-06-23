package types

import (
	mgl "github.com/go-gl/mathgl/mgl64"
	"image"
	"image/color"
	"math"
)

// FaceData represents a face in 3D space
type FaceData struct {
	Face         [3]mgl.Vec3 // The Face in 3D space as a list of vectors
	Color        color.Color // The Color of the Face
	Distance     Unit        // The Distance of the Face from the camera 3d world space
	TextureImage image.Image // The texture image for the face (nil if no texture)
	TexCoords    [3]mgl.Vec2 // Texture coordinates for each vertex
	HasTexture   bool        // Whether this face has texture information
	bounds       *AABB       // Cached bounds, nil if needs recalculation
	needsRecalc  bool        // Flag indicating if bounds need recalculation
}

// GetBounds returns the AABB for the face, recalculating if necessary
func (faceData *FaceData) GetBounds() AABB {
	if faceData.bounds == nil || faceData.needsRecalc {
		faceData.recalculateFaceBounds()
		faceData.needsRecalc = false
	}
	return *faceData.bounds
}

// recalculateFaceBounds computes the AABB for a face
func (faceData *FaceData) recalculateFaceBounds() {
	minPos := mgl.Vec3{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxPos := mgl.Vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}

	for _, vertex := range faceData.Face {
		minPos[0] = math.Min(minPos[0], vertex[0])
		minPos[1] = math.Min(minPos[1], vertex[1])
		minPos[2] = math.Min(minPos[2], vertex[2])
		maxPos[0] = math.Max(maxPos[0], vertex[0])
		maxPos[1] = math.Max(maxPos[1], vertex[1])
		maxPos[2] = math.Max(maxPos[2], vertex[2])
	}

	// Add small padding to avoid numerical precision issues
	padding := 0.01
	minPos = minPos.Sub(mgl.Vec3{padding, padding, padding})
	maxPos = maxPos.Add(mgl.Vec3{padding, padding, padding})

	if faceData.bounds == nil {
		faceData.bounds = &AABB{}
	}
	faceData.bounds.Min = minPos
	faceData.bounds.Max = maxPos
}

func (faceData *FaceData) Rotate(pivot mgl.Vec3, rotation mgl.Quat) {
	for i := 0; i < len(faceData.Face); i++ {
		p := faceData.Face[i].Sub(pivot)
		p = rotation.Rotate(p)
		faceData.Face[i] = p.Add(pivot)
	}
	faceData.needsRecalc = true
}

// Add adds another point to the face
func (faceData *FaceData) Add(other mgl.Vec3) {
	for i := 0; i < len(faceData.Face); i++ {
		faceData.Face[i] = faceData.Face[i].Add(other)
	}
	faceData.needsRecalc = true
}

// DistanceTo returns the distance between the face and a point
func (faceData *FaceData) DistanceTo(point mgl.Vec3) Unit {
	totalDistance := Unit(0)
	for _, vertex := range faceData.Face {
		totalDistance += Unit(vertex.Sub(point).Len())
	}
	return totalDistance / Unit(len(faceData.Face))
}

// Normal calculates the normal vector of the face
func (faceData *FaceData) Normal() mgl.Vec3 {
	edge1 := faceData.Face[1].Sub(faceData.Face[0])
	edge2 := faceData.Face[2].Sub(faceData.Face[0])
	normal := edge1.Cross(edge2).Normalize()
	return normal
}
