package renderer

import (
	"math"
	"sort"
)

func detectZBufferEdges(zBuffer [][]float64, baseThreshold, depthModulation, grazingAnglePower, grazingAngleHardness float64) [][]bool {
	w, h := len(zBuffer), len(zBuffer[0])
	mask := make([][]bool, w)
	for i := range mask {
		mask[i] = make([]bool, h)
	}
	var logzVals []float64
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			z := zBuffer[x][y]
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				logzVals = append(logzVals, math.Log(z))
			}
		}
	}
	if len(logzVals) == 0 {
		return mask
	}
	sort.Float64s(logzVals)
	lowIdx := int(0.02 * float64(len(logzVals)))
	highIdx := int(0.98 * float64(len(logzVals)))
	if highIdx <= lowIdx {
		lowIdx = 0
		highIdx = len(logzVals) - 1
	}
	minZ, maxZ := logzVals[lowIdx], logzVals[highIdx]
	if minZ == maxZ {
		minZ, maxZ = 0, 1
	}
	normZ := make([][]float64, w)
	for x := 0; x < w; x++ {
		normZ[x] = make([]float64, h)
		for y := 0; y < h; y++ {
			z := zBuffer[x][y]
			logz := maxZ
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				logz = math.Log(z)
			}
			normZ[x][y] = (logz - minZ) / (maxZ - minZ)
		}
	}
	gx := [3][3]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	gy := [3][3]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}
	for x := 1; x < w-1; x++ {
		for y := 1; y < h-1; y++ {
			var sx, sy float64
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					z := normZ[x+i][y+j]
					sx += gx[i+1][j+1] * z
					sy += gy[i+1][j+1] * z
				}
			}

			z := zBuffer[x][y]
			threshold := baseThreshold
			if !math.IsInf(z, 1) && !math.IsInf(z, -1) && z > 0 {
				threshold *= 1.0 + depthModulation*z
			}

			gradientMagnitude := math.Hypot(sx, sy)
			if gradientMagnitude > 0.001 {
				nx, ny := -sx/gradientMagnitude, -sy/gradientMagnitude

				nz := math.Sqrt(1.0 - nx*nx - ny*ny)
				if math.IsNaN(nz) {
					nz = 0.0
				}
				dotProduct := nz

				fresnel := math.Pow(1.0-math.Max(0.0, math.Min(1.0, dotProduct)), grazingAnglePower)
				grazingAngleMask := math.Max(0.0, math.Min(1.0, (fresnel+grazingAngleHardness-1.0)/grazingAngleHardness))
				threshold *= 1.0 / (1.0 + grazingAngleMask)
			}

			if gradientMagnitude > threshold {
				mask[x][y] = true
			}
		}
	}
	return mask
}
