# ThreeDView

A 3D widget for [Fyne](https://fyne.io) written in Go. 

## Overview

**ThreeDView** is a custom widget for the [Fyne](https://fyne.io) GUI toolkit, enabling developers to easily display and interact with 3D content in their Fyne applications. This library provides a simple API for integrating 3D visualizations with your desktop or mobile applications written in Go.

I built this because I wanted a 3D visualization in one of my other projects (https://github.com/virus-rpi/FlightControl) and because I wanted to learn more about 3D rendering.
This is compleatly written from scratch (somewhere in the development process I did start using [mathgl](github.com/go-gl/mathgl) to make sure I dont fuck up the math but if you check the git history I had my own implementations once).
I would not recommend using this in any bigger projects or anything where 3d is the main part of the application.
Because this is writting from scratch and in Go this can only do CPU rendering so you can't run too complex scenes. 
To still be as performant as possible I use smart go rotine creation to maximize CPU utilization (in tests i was able to achive up to 97% utilization).
The performance is heavily impacted by the resolution so phones are usually still able to render at a respectable fps because they have significantly smaller screen.  

I dont know if i will actually maintain this long term because this just started as a silly idea and escalated a bit to a point where it actually works as a 3d engine.

## Features

- Render 3D objects and scenes within Fyne apps
- Load .obj 3d files with textures or simplified colors from texture
- Interactive mouse/touch controls for rotation and zoom
- Extensible for custom geometries and camera controllers
- Manual and Orbit camera controller (orbit controller has a bug so currently i recomend implementing a custom controller)
- Pseudo lighting multiplying with the Z-Buffer
- Toggleable outline renderer for cartoony effect
- Seperate tick and render loop so animations are not affected by framerate
- Face outline renderer
- Wrieframe renderer
- Z-Buffer renderer
- Frustum culling to boost perfomance with oct-tree for fast frustum checks no matter how many objects there are
- Retrangulation of faces half outside the frustum and for models made out of non-triangle faces
- Simple geometric shape models included
- Ability to register any method into the tick loop
- Ability to limit TPS and FPS
- Ability to set a resolution factor to render at a smaller resolution than displayed for performance
- Easy way to create 3d models via code (look in object/models.go for examples)
- Cross platform (tested on Linux, Android and Windows 10)

## Screenshots

<img width="1048" height="808" alt="image" src="https://github.com/user-attachments/assets/f9919fb5-942d-4df7-8da3-8cbdaaebc9f6" />
15 fps with 126.2k faces (with a simpler model like the blender monkey i got up to 500fps)  

<img width="1047" height="802" alt="image" src="https://github.com/user-attachments/assets/50213984-cf43-4ff9-ba05-7f43fa9f502a" />
simplified color renderer instead of full texture  

<img width="1082" height="828" alt="image" src="https://github.com/user-attachments/assets/28c26b41-a7e5-4fe0-9a5c-08346f16a8ff" />
Z-Buffer renderer  

<img width="1082" height="933" alt="image" src="https://github.com/user-attachments/assets/06680343-c54a-47c4-aeaa-33e7c75ca8c3" />
Simpler model rendered at 60 fps with pseudo shading (limited to 60fps so my PC can do other things than just rendering)


## Installation

```bash
go get github.com/virus-rpi/ThreeDView
```

## Usage

Here’s a basic example of how to use ThreeDView in your Fyne application:

```go
package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/window"
    "github.com/virus-rpi/ThreeDView"
)

func main() {
    a := app.New()
    w := a.NewWindow("3D Demo")

    widget := ThreeDView.NewThreeDView()
    // Configure widget, add objects, etc.

    w.SetContent(widget)
    w.ShowAndRun()
}
```

## Documentation

For detailed usage, configuration options, and advanced features, see the [examples](./examples) directory and API comments in the code.  
You can also check my [other project](https://github.com/virus-rpi/FlightControl) (in the files simulation.go, rocket3d.go and control.go). 
It uses a significantly older version of the 3d engine so it might not work exactly like that anymore but this a example how a real usage of this could generally look like.

## Contributing

Pull requests, bug reports, and suggestions are welcome!
Especially if anyone can figure out why the oribt controller cant manage to point the camera at the target object (i tried litteraly everything)

## Known Bugs
- Broken orbit controller (no idea how to fix)
- One pixel gaps between some faces (will fix when i have time and/or motivation)

## Missing Features:

- **Lighting**: Currently there is no real lighting engine and currently I dont see a way to implement it with a usable performance except I somehow find a way to further optimize the current renderer
- **Custom cameras**: Currently you can only make custom camera controllers not custom cameras. I do plan to make the camera more modular so stuff like a orthographic camera becomes possible.
- **More exposed methods**: I want to some day add a lot more public methods to for example interact with the octree to make it possible to easily add colision or similar
- **Support non-triangular faces**: Currently my renderer can only render triangles. Models containing non-triangular faces currently just get re-meshed automatically but this adds more faces than nessesarry and therefore reducing performance

## License

This project is licensed under the MIT License. 
