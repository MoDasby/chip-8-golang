package main

import (
	"chip-8-golang/audio"
	"chip-8-golang/gfx"
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	boot()
	sdl.Init(sdl.INIT_AUDIO)
	beeper, err := audio.InitBeeper(10, 10)
	if err != nil {
		panic(err)
	}

	fmt.Println("Criando tela")
	window, renderer, err := gfx.CreateWindowAndRenderer(gfx.DEFAULT_SCALE, "Chip-8 emulator")
	if err != nil {
		panic(err)
	}
	fmt.Println("Tela criada")

	fmt.Println("Lendo instruções")
	for {
		event := sdl.PollEvent()

		if event != nil {
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

		for range 11 {
			instruction := uint16(memory[pc])<<8 | uint16(memory[pc+1]) // lê a próxima instrução de 2 bytes da memória
			pc += 2

			if instruction == 0xE0 {
				gfx.ClearBuffer(&screen)
				gfx.ClearRenderer(renderer)
				continue
			}

			process(instruction)
		}

		if delayTimer > 0 {
			delayTimer--
		}

		if soundTimer > 0 {
			fmt.Println("beep")
			beeper.Start()
			soundTimer--
		}

		if soundTimer <= 0 {
			beeper.Stop()
		}

		if drawFlag {
			gfx.Render(renderer, &screen, gfx.DEFAULT_SCALE)
			drawFlag = false
		}

		time.Sleep(time.Millisecond * 16)
	}
}
