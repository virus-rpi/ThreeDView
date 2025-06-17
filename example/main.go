package main

import (
	"ThreeDView"
	"ThreeDView/camera"
	"ThreeDView/object"
	"ThreeDView/types"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

	object.NewPlane(5000, types.Point3D{
		X: 0,
		Y: 0,
		Z: 0,
	},
		types.IdentityQuaternion(),
		color.RGBA{
			R: 0,
			G: 255,
			B: 0,
			A: 255,
		},
		threeDEnv,
		1)

	cube := object.NewCube(100, types.Point3D{
		X: 0,
		Y: 0,
		Z: 50,
	},
		types.IdentityQuaternion(),
		color.RGBA{
			R: 255,
			G: 0,
			B: 255,
			A: 255,
		}, threeDEnv)

	object.NewOrientationObject(threeDEnv)

	envCamera := camera.NewCamera(types.Point3D{
		X: 500,
		Y: 0,
		Z: 200,
	}, types.IdentityQuaternion())
	orbitController := camera.NewOrbitController(cube)
	envCamera.SetController(orbitController)
	threeDEnv.SetCamera(&envCamera)

	MainWindow.SetContent(threeDEnv)

	MainWindow.ShowAndRun()
}
