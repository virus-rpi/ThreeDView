package object

import (
	. "ThreeDView/camera"
	. "ThreeDView/types"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image/color"
	"sync"
)

type ThreeDWidgetInterface interface {
	GetCamera() *Camera
	RegisterTickMethod(func())
	AddObject(*Object)
	GetWidth() Pixel
	GetHeight() Pixel
}

// FaceData represents a face in 3D space
type FaceData struct {
	Face     Face        // The Face in 3D space as a Face
	Indices  [3]int      // The indices of the vertices in the original mesh
	Color    color.Color // The Color of the Face
	Distance Unit        // The Distance of the Face from the camera 3d world space
}

// ProjectedFaceData represents a face projected to 2D space
type ProjectedFaceData struct {
	Face     [3]mgl.Vec2 // The Face in 2D space as 3 2d points
	Z        [3]float64  // The Z (depth) value for each vertex
	Color    color.Color // The Color of the Face
	Distance Unit        // The Distance of the un-projected Face from the camera in 3d world space
}

// Object represents a 3D shape in world space
type Object struct {
	Faces    []FaceData            // Faces of the Object in local space
	Rotation mgl.Quat              // Rotation of the Object in world space (now quaternion)
	Position mgl.Vec3              // Position of the Object in world space
	Widget   ThreeDWidgetInterface // The Widget the Object is in
}

// GetFaces returns the faces of the shape in world space as FaceData
func (object *Object) GetFaces() []FaceData {
	faces := make([]FaceData, len(object.Faces))
	var wg sync.WaitGroup
	wg.Add(len(object.Faces))
	for i, face := range object.Faces {
		go func(i int, face FaceData) {
			defer wg.Done()
			clonedFace := face
			clonedFace.Face = face.Face
			clonedFace.Face.Rotate(mgl.Vec3{}, object.Rotation)
			clonedFace.Face.Add(object.Position)
			clonedFace.Distance = clonedFace.Face.DistanceTo(object.Widget.GetCamera().Position)
			faces[i] = clonedFace
		}(i, face)
	}
	wg.Wait()
	return faces
}

func (object *Object) GetPosition() mgl.Vec3 {
	return object.Position
}

func (object *Object) GetRotation() mgl.Quat {
	return object.Rotation
}

// GetEdges returns all unique edges of the object in world space as pairs of 3D points
func (object *Object) GetEdges() [][2]mgl.Vec3 {
	edgeMap := make(map[[2]int]struct{})
	var edges [][2]mgl.Vec3
	faces := object.GetFaces()
	for _, faceData := range faces {
		vertices := faceData.Face
		for i := 0; i < len(vertices); i++ {
			j := (i + 1) % len(vertices)
			idx1, idx2 := i, j
			if idx1 > idx2 {
				idx1, idx2 = idx2, idx1
			}
			edgeKey := [2]int{idx1, idx2}
			if _, exists := edgeMap[edgeKey]; !exists {
				edgeMap[edgeKey] = struct{}{}
				edges = append(edges, [2]mgl.Vec3{vertices[i], vertices[j]})
			}
		}
	}
	return edges
}

// FaceIndexData represents a face with both indices and world-space positions
type FaceIndexData struct {
	Indices [3]int
	World   [3]mgl.Vec3
	Normal  mgl.Vec3
	Front   bool
}

// GetSilhouetteEdges returns the silhouette (outer) edges of the object for the given camera position
func (object *Object) GetSilhouetteEdges(cameraPos mgl.Vec3) [][2]mgl.Vec3 {
	faces := object.GetFaces()

	type edgeKey struct {
		A, B mgl.Vec3
	}

	makeEdgeKey := func(a, b mgl.Vec3) edgeKey {
		if a.X() < b.X() || (a.X() == b.X() && (a.Y() < b.Y() || (a.Y() == b.Y() && a.Z() < b.Z()))) {
			return edgeKey{a, b}
		}
		return edgeKey{b, a}
	}

	edgeToFaceNormals := make(map[edgeKey][]mgl.Vec3)

	for _, faceData := range faces {
		vertices := faceData.Face
		normal := vertices[1].Sub(vertices[0]).Cross(vertices[2].Sub(vertices[0])).Normalize()
		for i := 0; i < len(vertices); i++ {
			j := (i + 1) % len(vertices)
			edge := makeEdgeKey(vertices[i], vertices[j])
			edgeToFaceNormals[edge] = append(edgeToFaceNormals[edge], normal)
		}
	}

	var silhouetteEdges [][2]mgl.Vec3
	for edge, normals := range edgeToFaceNormals {
		if len(normals) == 1 {
			silhouetteEdges = append(silhouetteEdges, [2]mgl.Vec3{edge.A, edge.B})
		} else if len(normals) == 2 {
			front1 := normals[0].Dot(cameraPos.Sub(edge.A).Normalize()) > 0
			front2 := normals[1].Dot(cameraPos.Sub(edge.A).Normalize()) > 0
			if front1 != front2 {
				silhouetteEdges = append(silhouetteEdges, [2]mgl.Vec3{edge.A, edge.B})
			}
		}
	}
	return silhouetteEdges
}
