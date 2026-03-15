package shaders

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed crt.kage
var crtShaderSrc []byte

var globalCRTShader *ebiten.Shader

func GetCRTShader() *ebiten.Shader {
	if globalCRTShader == nil {
		s, err := ebiten.NewShader(crtShaderSrc)
		if err != nil {
			return nil
		}
		globalCRTShader = s
	}
	return globalCRTShader
}
