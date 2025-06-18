package types

import "image/color"

type ThreeDWidgetInterface interface {
	RegisterTickMethod(func())
	GetWidth() Pixel
	GetHeight() Pixel
	GetBackgroundColor() color.Color
	GetRenderFaceColors() bool
	GetRenderTextures() bool
	GetRenderFaceOutlines() bool
	GetRenderEdgeOutlines() bool
	GetRenderZBuffer() bool
}
