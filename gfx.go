package main

import (
	"chip-8-golang/shaders"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenW = 64
	ScreenH = 32
	Scale   = 20
)

var KeyMap = map[ebiten.Key]uint16{
	ebiten.Key1: 0x1, ebiten.Key2: 0x2, ebiten.Key3: 0x3, ebiten.Key4: 0xC,
	ebiten.KeyQ: 0x4, ebiten.KeyW: 0x5, ebiten.KeyE: 0x6, ebiten.KeyR: 0xD,
	ebiten.KeyA: 0x7, ebiten.KeyS: 0x8, ebiten.KeyD: 0x9, ebiten.KeyF: 0xE,
	ebiten.KeyZ: 0xA, ebiten.KeyX: 0x0, ebiten.KeyC: 0xB, ebiten.KeyV: 0xF,
}

type GFX struct {
	PrimaryColor   color.RGBA
	SecondaryColor color.RGBA
	cpu            *CPU
	audioPlayer    *audio.Player
	offscreen      *ebiten.Image
	crtTime        float64
}

func (g *GFX) Update() error {
	for key, hexCode := range KeyMap {
		g.cpu.updateKeyboard(hexCode, ebiten.IsKeyPressed(key))
	}

	g.cpu.fetch()

	if g.cpu.soundTimer > 0 {
		if !g.audioPlayer.IsPlaying() {
			g.audioPlayer.Play()
		}
	} else {
		if g.audioPlayer.IsPlaying() {
			g.audioPlayer.Pause()
		}
	}

	return nil
}

func (g *GFX) Draw(screenImage *ebiten.Image) {
	sw, sh := ScreenW*Scale, ScreenH*Scale

	if g.offscreen == nil {
		g.offscreen = ebiten.NewImage(sw, sh)
	}

	g.offscreen.Fill(g.PrimaryColor)
	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			if g.cpu.screen[y][x] == 1 {
				vector.FillRect(g.offscreen,
					float32(x*Scale), float32(y*Scale),
					float32(Scale), float32(Scale),
					g.SecondaryColor, false)
			}
		}
	}

	shader := shaders.GetCRTShader()
	if shader != nil {
		g.crtTime += 1.0 / 60.0
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = g.offscreen
		op.Uniforms = map[string]any{
			"Time":             float32(g.crtTime),
			"PixelScale":       float32(Scale),
			"Brightness":       float32(1.0),
			"ScanlineStrength": float32(1.0),
			"GridStrength":     float32(1.0),
			"VignetteStrength": float32(1.0),
			"TintStrength":     float32(1.0),
		}
		screenImage.DrawRectShader(sw, sh, shader, op)
		return
	}

	screenImage.DrawImage(g.offscreen, nil)
}

func (g *GFX) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenW * Scale, ScreenH * Scale
}
