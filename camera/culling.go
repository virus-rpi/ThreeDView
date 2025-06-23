package camera

import (
	"ThreeDView/types"
	mgl "github.com/go-gl/mathgl/mgl64"
	"sync"
)

type octreeNode struct {
	Bounds   types.AABB
	Depth    int
	MaxDepth int
	MaxItems int
	Children []*octreeNode
	Faces    []types.FaceData
	Parent   *octreeNode
	sync.RWMutex
}

func newOctree(bounds types.AABB, maxDepth, maxItems int) *octreeNode {
	return &octreeNode{
		Bounds:   bounds,
		MaxDepth: maxDepth,
		MaxItems: maxItems,
		Children: make([]*octreeNode, 8), // allocate space for 8 children
	}
}

func (n *octreeNode) insert(face types.FaceData) {
	n.Lock()
	defer n.Unlock()

	if n.Depth == n.MaxDepth {
		n.Faces = append(n.Faces, face)
		return
	}

	if len(n.Faces) < n.MaxItems && len(n.Children) == 0 {
		n.Faces = append(n.Faces, face)
		return
	}

	// Split if not already split
	if n.Children[0] == nil {
		n.split()
	}

	// Try to insert into a child
	for _, child := range n.Children {
		if child.Bounds.Contains(face.GetBounds()) {
			child.insert(face)
			return
		}
	}

	// If it doesn't fit neatly into any child
	n.Faces = append(n.Faces, face)
}

func (n *octreeNode) split() {
	center := n.Bounds.Center()

	for i := 0; i < 8; i++ {
		newMin := n.Bounds.Min
		newMax := center

		if i&1 != 0 {
			newMin[0] = center[0]
			newMax[0] = n.Bounds.Max[0]
		}
		if i&2 != 0 {
			newMin[1] = center[1]
			newMax[1] = n.Bounds.Max[1]
		}
		if i&4 != 0 {
			newMin[2] = center[2]
			newMax[2] = n.Bounds.Max[2]
		}

		n.Children[i] = &octreeNode{
			Bounds:   types.AABB{Min: newMin, Max: newMax},
			Depth:    n.Depth + 1,
			MaxDepth: n.MaxDepth,
			MaxItems: n.MaxItems,
			Parent:   n,
			Children: make([]*octreeNode, 8),
		}
	}

	// Re-distribute existing faces
	oldFaces := n.Faces
	n.Faces = nil
	for _, face := range oldFaces {
		inserted := false
		for _, child := range n.Children {
			if child.Bounds.Contains(face.GetBounds()) {
				child.insert(face)
				inserted = true
				break
			}
		}
		if !inserted {
			n.Faces = append(n.Faces, face)
		}
	}
}

func (n *octreeNode) query(frustum Frustum, callbackChan chan types.FaceData, wg *sync.WaitGroup) {
	defer wg.Done()
	n.RLock()
	defer n.RUnlock()

	if !frustum.Intersects(n.Bounds) {
		return
	}

	for _, face := range n.Faces {
		if frustum.Intersects(face.GetBounds()) {
			callbackChan <- face
		}
	}

	if n.Children[0] != nil {
		for _, child := range n.Children {
			wg.Add(1)
			child.query(frustum, callbackChan, wg)
		}
	}
}

type Frustum struct {
	Planes [6]Plane
}

type Plane struct {
	Normal mgl.Vec3
	D      float64
}

// Intersects AABB-Frustum intersection test
func (f *Frustum) Intersects(box types.AABB) bool {
	for _, plane := range f.Planes {
		px := box.Min.X()
		py := box.Min.Y()
		pz := box.Min.Z()

		if plane.Normal.X() >= 0 {
			px = box.Max.X()
		}
		if plane.Normal.Y() >= 0 {
			py = box.Max.Y()
		}
		if plane.Normal.Z() >= 0 {
			pz = box.Max.Z()
		}

		if plane.Normal.Dot(mgl.Vec3{px, py, pz})+plane.D < 0 {
			return false
		}
	}
	return true
}
