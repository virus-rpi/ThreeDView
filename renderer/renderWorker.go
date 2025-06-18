package renderer

import (
	"ThreeDView/object"
	"ThreeDView/types"
	mgl "github.com/go-gl/mathgl/mgl64"
	"log"
)

type instruction struct {
	instructionType string
	data            interface{}
	callbackChannel chan interface{}
	doneFunction    func()
}

type renderWorker struct {
	w               threeDWidgetInterface
	workerChannel   chan *instruction
	shouldTerminate bool
	State           string
	Success         int
}

func newRenderWorker(w threeDWidgetInterface, workerChannel chan *instruction) *renderWorker {
	return &renderWorker{w: w, workerChannel: workerChannel, shouldTerminate: false, State: "initialized", Success: 0}
}

func (rw *renderWorker) start() {
	rw.State = "idle"
	for newInstruction := range rw.workerChannel {
		rw.processInstruction(newInstruction)

		if rw.shouldTerminate {
			log.Println("Render worker terminating...")
			return
		}
	}
}

func (rw *renderWorker) processInstruction(instruction *instruction) {
	rw.State = instruction.instructionType
	switch instruction.instructionType {
	case "terminate":
		rw.shouldTerminate = true
		break
	case "getFacesInFrustum":
		rw.getFacesInFrustum(instruction)
		break
	case "clipAndProject":
		rw.clipAndProject(instruction)
		break
	}
	rw.State = "done"
	instruction.doneFunction()
	rw.State = "idle"
	rw.Success++
}

func (rw *renderWorker) getFacesInFrustum(instruction *instruction) {
	for _, face := range instruction.data.([]object.FaceData) {
		if rw.w.GetCamera().FaceOverlapsFrustum(face.Face, rw.w.GetWidth(), rw.w.GetHeight()) {
			instruction.callbackChannel <- face
		}
	}
}

func (rw *renderWorker) clipAndProject(instruction *instruction) {
	face := instruction.data.(object.FaceData)

	clippedPolys := rw.w.GetCamera().ClipAndProjectFace(face.Face, rw.w.GetWidth(), rw.w.GetHeight(), face.TexCoords)
	if clippedPolys == nil {
		return
	}

	for _, triangle := range clippedPolys {
		if !triangleOverlapsScreen(triangle.Points[0], triangle.Points[1], triangle.Points[2], rw.w.GetWidth(), rw.w.GetHeight()) {
			continue
		}

		projectedFace := object.ProjectedFaceData{
			Face:     triangle.Points,
			Z:        triangle.Z,
			Color:    face.Color,
			Distance: face.Distance,
		}

		if face.HasTexture && triangle.HasTexture {
			projectedFace.TextureImage = face.TextureImage
			projectedFace.TexCoords = triangle.TexCoords
			projectedFace.HasTexture = true
		}

		instruction.callbackChannel <- projectedFace
	}
}

func triangleOverlapsScreen(p1, p2, p3 mgl.Vec2, width, height types.Pixel) bool {
	minX := min(int(p1.X()), min(int(p2.X()), int(p3.X())))
	maxX := max(int(p1.X()), max(int(p2.X()), int(p3.X())))
	minY := min(int(p1.Y()), min(int(p2.Y()), int(p3.Y())))
	maxY := max(int(p1.Y()), max(int(p2.Y()), int(p3.Y())))
	return maxX >= 0 && minX < int(width) && maxY >= 0 && minY < int(height)
}
