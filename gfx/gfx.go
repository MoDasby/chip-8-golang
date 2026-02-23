package gfx

import "github.com/veandco/go-sdl2/sdl"

const (
	ScreenW       = 64
	ScreenH       = 32
	DEFAULT_SCALE = 20
)

var KEYBOARD_MAP = map[sdl.Keycode]uint16{
	sdl.K_1: 0x1,
	sdl.K_2: 0x2,
	sdl.K_3: 0x3,
	sdl.K_4: 0xC,
	sdl.K_q: 0x4,
	sdl.K_w: 0x5,
	sdl.K_e: 0x6,
	sdl.K_r: 0xD,
	sdl.K_a: 0x7,
	sdl.K_s: 0x8,
	sdl.K_d: 0x9,
	sdl.K_f: 0xE,
	sdl.K_z: 0xA,
	sdl.K_x: 0x0,
	sdl.K_c: 0xB,
	sdl.K_v: 0xF,
}

func CreateWindowAndRenderer(scale int32, title string) (*sdl.Window, *sdl.Renderer, error) {
	window, err := sdl.CreateWindow(
		title,
		int32(sdl.WINDOWPOS_CENTERED),
		int32(sdl.WINDOWPOS_CENTERED),
		ScreenW*scale,
		ScreenH*scale,
		uint32(sdl.WINDOW_SHOWN),
	)
	if err != nil {
		return nil, nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, uint32(sdl.RENDERER_ACCELERATED))
	if err != nil {
		window.Destroy()
		return nil, nil, err
	}

	return window, renderer, nil
}

func ClearBuffer(screen *[ScreenH][ScreenW]byte) {
	for y := range ScreenH {
		for x := 0; x < ScreenW; x++ {
			screen[y][x] = 0
		}
	}
}

func ClearRenderer(r *sdl.Renderer) error {
	if err := r.SetDrawColor(15, 56, 15, 255); err != nil {
		return err
	}
	return r.Clear()
}

func Render(r *sdl.Renderer, screen *[ScreenH][ScreenW]byte, scale int32) error {
	if err := ClearRenderer(r); err != nil {
		return err
	}

	if err := r.SetDrawColor(139, 172, 15, 255); err != nil {
		return err
	}

	for y := range ScreenH {
		for x := range ScreenW {
			if screen[y][x] == 1 {
				rect := sdl.Rect{
					X: int32(x) * scale,
					Y: int32(y) * scale,
					W: scale,
					H: scale,
				}
				if err := r.FillRect(&rect); err != nil {
					return err
				}
			}
		}
	}

	r.Present()
	return nil
}
