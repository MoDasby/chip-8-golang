package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
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
}

func (g *GFX) Update() error {
	for key, hexCode := range KeyMap {
		g.cpu.updateKeyboard(hexCode, ebiten.IsKeyPressed(key))
	}

	g.cpu.fetch()

	return nil
}

func (g *GFX) Draw(screenImage *ebiten.Image) {
	screenImage.Fill(g.PrimaryColor)

	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			if g.cpu.screen[y][x] == 1 {
				vector.FillRect(screenImage, float32(x), float32(y), 1, 1, g.SecondaryColor, false)
			}
		}
	}
}

func (g *GFX) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenW, ScreenH
}
