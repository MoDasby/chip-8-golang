package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

func convertRGB(strs []string) ([]uint8, error) {
	// Expect exactly 3 components: R, G, B.
	if len(strs) != 3 {
		return nil, fmt.Errorf("invalid RGB value %v: expected 3 components, got %d", strs, len(strs))
	}

	result := make([]uint8, len(strs))

	for i, s := range strs {
		v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid RGB component %q: %w", s, err)
		}
		result[i] = uint8(v)
	}

	return result, nil
}

func main() {
	cpu := NewCpu()

	choosedGame, err := boot(cpu)
	if err != nil {
		panic(err)
	}

	primaryRGB, _ := convertRGB(strings.Split(choosedGame.Theme.PrimaryColor, ","))
	secondaryRGB, _ := convertRGB(strings.Split(choosedGame.Theme.SecondaryColor, ","))

	audioContext := audio.NewContext(sampleRate)

	player, err := audioContext.NewPlayer(&beepStream{})
	if err != nil {
		panic(err)
	}

	game := &GFX{
		PrimaryColor:   color.RGBA{R: primaryRGB[0], G: primaryRGB[1], B: primaryRGB[2], A: 255},
		SecondaryColor: color.RGBA{R: secondaryRGB[0], G: secondaryRGB[1], B: secondaryRGB[2], A: 255},
		cpu:            cpu,
		audioPlayer:    player,
	}

	ebiten.SetWindowSize(ScreenW*Scale, ScreenH*Scale)
	ebiten.SetWindowTitle(choosedGame.Name)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
