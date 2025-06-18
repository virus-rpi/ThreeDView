package renderer

import (
	"ThreeDView/object"
	. "ThreeDView/types"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image"
	"image/color"
	"math"
)

func drawFilledTriangle(img *image.RGBA, face object.ProjectedFaceData, zBuffer [][]float64, useTexture bool) {
	p := face.Face
	z := face.Z
	fill := face.Color
	textureImg := face.TextureImage
	texCoords := face.TexCoords
	v := [3]struct {
		p mgl.Vec2
		z float64
		t mgl.Vec2
	}{{p[0], z[0], texCoords[0]}, {p[1], z[1], texCoords[1]}, {p[2], z[2], texCoords[2]}}

	if v[1].p.Y() < v[0].p.Y() {
		v[0], v[1] = v[1], v[0]
	}
	if v[2].p.Y() < v[0].p.Y() {
		v[0], v[2] = v[2], v[0]
	}
	if v[2].p.Y() < v[1].p.Y() {
		v[1], v[2] = v[2], v[1]
	}

	interpolate := func(y, y1, y2, x1, x2, z1, z2 float64, t1, t2 mgl.Vec2) (Pixel, float64, mgl.Vec2) {
		if y1 == y2 {
			return Pixel(x1), z1, t1
		}
		t := (y - y1) / (y2 - y1)
		return Pixel(x1 + (x2-x1)*t),
			z1 + (z2-z1)*t,
			mgl.Vec2{t1.X() + (t2.X()-t1.X())*t, t1.Y() + (t2.Y()-t1.Y())*t}
	}

	getTextureColor := func(texImg image.Image, texCoord mgl.Vec2) color.Color {
		if texImg == nil {
			return fill
		}

		bounds := texImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		x := int(math.Floor(texCoord.X() * float64(width-1)))
		y := int(math.Floor((1.0 - texCoord.Y()) * float64(height-1))) // Flip Y coordinate

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

		return texImg.At(x, y)
	}

	for yf := math.Ceil(v[0].p.Y()); yf <= v[1].p.Y(); yf++ {
		x1, z1, t1 := interpolate(yf, v[0].p.Y(), v[1].p.Y(), v[0].p.X(), v[1].p.X(), v[0].z, v[1].z, v[0].t, v[1].t)
		x2, z2, t2 := interpolate(yf, v[0].p.Y(), v[2].p.Y(), v[0].p.X(), v[2].p.X(), v[0].z, v[2].z, v[0].t, v[2].t)
		if x1 > x2 {
			x1, x2, z1, z2, t1, t2 = x2, x1, z2, z1, t2, t1
		}
		for x := int(math.Ceil(float64(x1))); float64(x) <= float64(x2); x++ {
			if x >= 0 && x < img.Bounds().Dx() && int(yf) >= 0 && int(yf) < img.Bounds().Dy() {
				t := 0.0
				if x2 != x1 {
					t = float64(x-int(x1)) / float64(x2-x1)
				}
				z := z1 + (z2-z1)*t
				texCoord := mgl.Vec2{t1.X() + (t2.X()-t1.X())*t, t1.Y() + (t2.Y()-t1.Y())*t}

				if x >= len(zBuffer) || int(yf) >= len(zBuffer[x]) {
					continue
				}
				if z < zBuffer[x][int(yf)] {
					zBuffer[x][int(yf)] = z

					// Use texture if available and enabled
					if useTexture && textureImg != nil {
						img.Set(x, int(yf), getTextureColor(textureImg, texCoord))
					} else {
						img.Set(x, int(yf), fill)
					}
				}
			}
		}
	}

	for yf := v[1].p.Y(); yf <= v[2].p.Y(); yf++ {
		x1, z1, t1 := interpolate(yf, v[1].p.Y(), v[2].p.Y(), v[1].p.X(), v[2].p.X(), v[1].z, v[2].z, v[1].t, v[2].t)
		x2, z2, t2 := interpolate(yf, v[0].p.Y(), v[2].p.Y(), v[0].p.X(), v[2].p.X(), v[0].z, v[2].z, v[0].t, v[2].t)
		if x1 > x2 {
			x1, x2, z1, z2, t1, t2 = x2, x1, z2, z1, t2, t1
		}
		for x := int(math.Ceil(float64(x1))); float64(x) <= float64(x2); x++ {
			if x >= 0 && x < img.Bounds().Dx() && int(yf) >= 0 && int(yf) < img.Bounds().Dy() {
				t := 0.0
				if x2 != x1 {
					t = float64(x-int(x1)) / float64(x2-x1)
				}
				z := z1 + (z2-z1)*t
				texCoord := mgl.Vec2{t1.X() + (t2.X()-t1.X())*t, t1.Y() + (t2.Y()-t1.Y())*t}

				if x >= len(zBuffer) || int(yf) >= len(zBuffer[x]) {
					continue
				}
				if z < zBuffer[x][int(yf)] {
					zBuffer[x][int(yf)] = z

					if useTexture && textureImg != nil {
						img.Set(x, int(yf), getTextureColor(textureImg, texCoord))
					} else {
						img.Set(x, int(yf), fill)
					}
				}
			}
		}
	}
}
