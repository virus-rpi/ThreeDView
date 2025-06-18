package main

import (
	"ThreeDView"
	"ThreeDView/camera"
	"ThreeDView/object"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image/color"
)

func main() {
	App := app.New()
	MainWindow := App.NewWindow("3D View Example")
	MainWindow.Resize(fyne.NewSize(800, 600))
	MainWindow.CenterOnScreen()

	threeDEnv := ThreeDView.NewThreeDWidget()
	threeDEnv.SetBackgroundColor(color.RGBA{
		R: 135,
		G: 206,
		B: 235,
		A: 255,
	})
	threeDEnv.SetResolutionFactor(1.0)

	center, _ := object.NewObjectFromObjFile("./example/assets/stress-boat.obj", mgl.Vec3{0, 100, 0}, mgl.QuatIdent(), 100, color.RGBA{R: 255, B: 255, A: 255}, "./example/assets/stress-boat-texture.jpg", threeDEnv)

	object.NewOrientationObject(threeDEnv)

	envCamera := camera.NewCamera(mgl.Vec3{
		500,
		0,
		200,
	}, mgl.QuatIdent())
	// manualController := camera.NewManualController()
	// manualController.ShowControlWindow()
	// envCamera.SetController(manualController)
	orbitController := camera.NewOrbitController(center)
	envCamera.SetController(orbitController)
	threeDEnv.SetCamera(&envCamera)

	MainWindow.SetContent(threeDEnv)

	controlWindow := App.NewWindow("3D Debug Controls")
	controlWindow.Resize(fyne.NewSize(300, 200))

	zBufferCheck := widget.NewCheck("Show Z Buffer Debug", func(checked bool) {
		threeDEnv.SetRenderZBufferDebug(checked)
	})

	edgeCheck := widget.NewCheck("Show Edge Outline", func(checked bool) {
		threeDEnv.SetRenderEdgeOutline(checked)
	})
	edgeCheck.SetChecked(false)

	faceOutlineCheck := widget.NewCheck("Show Face Outlines", func(checked bool) {
		threeDEnv.SetRenderFaceOutlines(checked)
	})
	faceOutlineCheck.SetChecked(false)

	faceColorCheck := widget.NewCheck("Show Face Colors", func(checked bool) {
		threeDEnv.SetRenderFaceColors(checked)
	})
	faceColorCheck.SetChecked(true)

	textureCheck := widget.NewCheck("Use Textures", func(checked bool) {
		threeDEnv.SetRenderTextures(checked)
	})
	textureCheck.SetChecked(true)

	controls := container.New(
		layout.NewVBoxLayout(),
		zBufferCheck,
		edgeCheck,
		faceOutlineCheck,
		faceColorCheck,
		textureCheck,
	)
	controlWindow.SetContent(controls)
	controlWindow.Show()

	MainWindow.ShowAndRun()
}
