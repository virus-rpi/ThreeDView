package object

import (
	"bufio"
	"fmt"
	"github.com/flywave/go-earcut"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// NewObjectFromObjFile parses a Wavefront OBJ file at 'path', triangulates all faces
func NewObjectFromObjFile(path string, position mgl.Vec3, rotation mgl.Quat, scale float64, col color.Color, w ThreeDWidgetInterface) (*Object, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open OBJ file: %v", err)
	}
	defer file.Close()

	var vertices []mgl.Vec3
	var facesData []FaceData

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}

		switch tokens[0] {
		case "v":
			if len(tokens) < 4 {
				continue
			}
			x, _ := strconv.ParseFloat(tokens[1], 64)
			y, _ := strconv.ParseFloat(tokens[2], 64)
			z, _ := strconv.ParseFloat(tokens[3], 64)
			vertices = append(vertices, mgl.Vec3{x * scale, y * scale, z * scale})

		case "f":
			var faceVertices []float64
			var vertexIndices []int
			for _, tok := range tokens[1:] {
				parts := strings.Split(tok, "/")
				i, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}
				if i < 0 {
					i = len(vertices) + i
				} else {
					i--
				}
				vertexIndices = append(vertexIndices, i)
				v := vertices[i]
				faceVertices = append(faceVertices, v[0], v[1], v[2])
			}

			// Triangulate the face using earcut
			var holes []int // No holes in this case
			dim := 3        // 3D coordinates
			triangles, _ := earcut.Earcut(faceVertices, holes, dim)

			// Create triangles from the indices
			for i := 0; i < len(triangles); i += 3 {
				idx1 := triangles[i]
				idx2 := triangles[i+1]
				idx3 := triangles[i+2]

				// Convert back to original vertex indices
				v1 := vertices[vertexIndices[idx1]]
				v2 := vertices[vertexIndices[idx2]]
				v3 := vertices[vertexIndices[idx3]]

				facesData = append(facesData, FaceData{
					Face:  [3]mgl.Vec3{v1, v2, v3},
					Color: col,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading OBJ file: %v", err)
	}

	obj := &Object{
		Faces:    facesData,
		Position: position,
		Rotation: rotation,
		Widget:   w,
	}
	w.AddObject(obj)
	return obj, nil
}
