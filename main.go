package main

import (
	"chip-8-golang/gfx"
	"strconv"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func convertRGB(strs []string) []uint8 {
	result := make([]uint8, len(strs))

	for i, s := range strs {
		v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 8)
		if err != nil {
			panic(err)
		}
		result[i] = uint8(v)
	}

	return result
}

func main() {

	choosedGame, err := boot()
	if err != nil {
		panic(err)
	}

	sdl.Init(sdl.INIT_AUDIO)

	window, renderer, err := gfx.CreateWindowAndRenderer(
		gfx.DEFAULT_SCALE,
		choosedGame.Name,
		convertRGB(strings.Split(choosedGame.Theme.PrimaryColor, ",")),
		convertRGB(strings.Split(choosedGame.Theme.SecondaryColor, ",")),
	)
	if err != nil {
		panic(err)
	}

	var dt time.Duration
	last := time.Now()

	var cpuAcc time.Duration = 0
	cpuStep := time.Second / 700

	var timerAcc time.Duration = 0
	timerStep := time.Second / 60

	var renderAcc time.Duration = 0
	renderStep := time.Second / 60

	for {
		now := time.Now()
		dt = now.Sub(last)
		last = now

		cpuAcc += dt
		timerAcc += dt
		renderAcc += dt

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				window.Destroy()
				return
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN && e.Repeat == 0 {
					keyboard[gfx.KEYBOARD_MAP[e.Keysym.Sym]] = true
				}

				if e.Type == sdl.KEYUP {
					keyboard[gfx.KEYBOARD_MAP[e.Keysym.Sym]] = false
				}
			}
		}

		for cpuAcc >= cpuStep {
			instruction := uint16(memory[pc])<<8 | uint16(memory[pc+1]) // lê a próxima instrução de 2 bytes da memória
			pc += 2

			if instruction == 0xE0 {
				gfx.ClearBuffer(&screen)
				gfx.ClearRenderer(renderer)

				cpuAcc -= cpuStep
				continue
			}

			process(instruction)

			cpuAcc -= cpuStep
		}

		for timerAcc >= timerStep {
			if delayTimer > 0 {
				delayTimer--
			}

			if soundTimer > 0 {
				soundTimer--
			}

			timerAcc -= timerStep
		}

		for renderAcc >= renderStep {
			if drawFlag {
				gfx.Render(renderer, &screen, gfx.DEFAULT_SCALE)
			}

			renderAcc -= renderStep
		}
	}
}
