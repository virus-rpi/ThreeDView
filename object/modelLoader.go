package object

import (
	"bufio"
	"fmt"
	"github.com/flywave/go-earcut"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strconv"
	"strings"
)

// getAverageColorFromTexture calculates the average color from a texture using the provided texture coordinates
func getAverageColorFromTexture(img image.Image, texCoords []mgl.Vec2) color.Color {
	if len(texCoords) == 0 {
		return color.RGBA{A: 255}
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var totalR, totalG, totalB, totalA uint32
	count := 0

	for _, tc := range texCoords {
		x := int(math.Floor(tc[0] * float64(width-1)))
		y := int(math.Floor((1.0 - tc[1]) * float64(height-1)))

		if x < 0 {
			x = 0
		} else if x >= width {
			x = width - 1
		}
		if y < 0 {
			y = 0
		} else if y >= height {
			y = height - 1
		}

		r, g, b, a := img.At(x, y).RGBA()
		totalR += r
		totalG += g
		totalB += b
		totalA += a
		count++
	}

	if count == 0 {
		return color.RGBA{A: 255}
	}

	return color.RGBA{
		R: uint8(totalR / uint32(count) >> 8),
		G: uint8(totalG / uint32(count) >> 8),
		B: uint8(totalB / uint32(count) >> 8),
		A: uint8(totalA / uint32(count) >> 8),
	}
}

// NewObjectFromObjFile parses a Wavefront OBJ file at 'path', triangulates all faces
// If texturePath is provided, it will use the texture to determine face colors
func NewObjectFromObjFile(path string, position mgl.Vec3, rotation mgl.Quat, scale float64, col color.Color, texturePath string, w threeDWidgetInterface) (*Object, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open OBJ file: %v", err)
	}
	defer file.Close()

	var vertices []mgl.Vec3
	var texCoords []mgl.Vec2
	var facesData []FaceData

	var textureImg image.Image
	if texturePath != "" {
		textureFile, err := os.Open(texturePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open texture file: %v", err)
		}
		defer textureFile.Close()

		textureImg, _, err = image.Decode(textureFile)
		if err != nil {
			return nil, fmt.Errorf("failed to decode texture image: %v", err)
		}
	}

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

		case "vt":
			if len(tokens) < 3 {
				continue
			}
			u, _ := strconv.ParseFloat(tokens[1], 64)
			v, _ := strconv.ParseFloat(tokens[2], 64)
			texCoords = append(texCoords, mgl.Vec2{u, v})

		case "f":
			var faceVertices []float64
			var vertexIndices []int
			var texCoordIndices []int

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

				if len(parts) > 1 && parts[1] != "" && texturePath != "" {
					ti, err := strconv.Atoi(parts[1])
					if err == nil {
						if ti < 0 {
							ti = len(texCoords) + ti
						} else {
							ti--
						}
						if ti >= 0 && ti < len(texCoords) {
							texCoordIndices = append(texCoordIndices, ti)
						}
					}
				}
			}

			var holes []int
			dim := 3
			triangles, _ := earcut.Earcut(faceVertices, holes, dim)

			for i := 0; i < len(triangles); i += 3 {
				idx1 := triangles[i]
				idx2 := triangles[i+1]
				idx3 := triangles[i+2]

				v1 := vertices[vertexIndices[idx1]]
				v2 := vertices[vertexIndices[idx2]]
				v3 := vertices[vertexIndices[idx3]]

				faceColor := col

				// Create face data
				faceData := FaceData{
					Face:  [3]mgl.Vec3{v1, v2, v3},
					Color: faceColor,
				}

				// Handle texture if available
				if textureImg != nil && len(texCoordIndices) >= 3 {
					var triangleTexCoords []mgl.Vec2
					if idx1 < len(texCoordIndices) {
						triangleTexCoords = append(triangleTexCoords, texCoords[texCoordIndices[idx1]])
					}
					if idx2 < len(texCoordIndices) {
						triangleTexCoords = append(triangleTexCoords, texCoords[texCoordIndices[idx2]])
					}
					if idx3 < len(texCoordIndices) {
						triangleTexCoords = append(triangleTexCoords, texCoords[texCoordIndices[idx3]])
					}

					if len(triangleTexCoords) > 0 {
						// Set the average color from texture
						faceData.Color = getAverageColorFromTexture(textureImg, triangleTexCoords)

						// If we have exactly 3 texture coordinates, store them for texture mapping
						if len(triangleTexCoords) == 3 {
							faceData.TextureImage = textureImg
							faceData.TexCoords = [3]mgl.Vec2{
								triangleTexCoords[0],
								triangleTexCoords[1],
								triangleTexCoords[2],
							}
							faceData.HasTexture = true
						}
					}
				}

				facesData = append(facesData, faceData)
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
