package main

import (
	"ThreeDView"
	"ThreeDView/camera"
	"ThreeDView/object"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	mgl "github.com/go-gl/mathgl/mgl64"
	"image/color"
)

func main() {
	App := app.New()
	MainWindow := App.NewWindow("3D View Example")
	MainWindow.Resize(fyne.NewSize(800, 600))
	MainWindow.CenterOnScreen()

	threeDEnv := ThreeDView.NewThreeDWidget()
	threeDEnv.Hide()
	threeDEnv.SetBackgroundColor(color.RGBA{
		R: 135,
		G: 206,
		B: 235,
		A: 255,
	})
	threeDEnv.SetResolutionFactor(1.0)
	threeDEnv.SetRenderFaceOutlines(true)

	object.NewPlane(5000, mgl.Vec3{
		0,
		0,
		0,
	},
		mgl.QuatIdent(),
		color.RGBA{
			R: 0,
			G: 255,
			B: 0,
			A: 255,
		},
		threeDEnv,
		1)

	cube := object.NewCube(100, mgl.Vec3{
		0,
		0,
		50,
	},
		mgl.QuatIdent(),
		color.RGBA{
			R: 255,
			G: 0,
			B: 255,
			A: 255,
		}, threeDEnv)

	object.NewOrientationObject(threeDEnv)

	envCamera := camera.NewCamera(mgl.Vec3{
		500,
		0,
		200,
	}, mgl.QuatIdent())
	orbitController := camera.NewOrbitController(cube)
	envCamera.SetController(orbitController)
	threeDEnv.SetCamera(&envCamera)

	MainWindow.SetContent(threeDEnv)

	MainWindow.ShowAndRun()
}
