package main

import (
	"flag"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

type AppState int

const (
	AppStateMenu AppState = iota
	AppStatePlaying
)

type App struct {
	state    AppState
	menu     *Menu
	gfx      *GFX
	audioCtx *audio.Context
}

func (a *App) launchGame(game *GameMetadata) error {
	cpu := NewCpu()
	cpu.LoadFontset()

	if err := cpu.LoadROM(normalizeRomLocation(game)); err != nil {
		return err
	}

	primaryRGB, _ := convertRGB(strings.Split(game.Theme.PrimaryColor, ","))
	secondaryRGB, _ := convertRGB(strings.Split(game.Theme.SecondaryColor, ","))

	player, err := a.audioCtx.NewPlayer(&beepStream{})
	if err != nil {
		return err
	}

	a.gfx = &GFX{
		PrimaryColor:   color.RGBA{R: primaryRGB[0], G: primaryRGB[1], B: primaryRGB[2], A: 255},
		SecondaryColor: color.RGBA{R: secondaryRGB[0], G: secondaryRGB[1], B: secondaryRGB[2], A: 255},
		cpu:            cpu,
		audioPlayer:    player,
	}

	ebiten.SetWindowSize(ScreenW*Scale, ScreenH*Scale)
	ebiten.SetWindowTitle(game.Name)
	a.state = AppStatePlaying
	return nil
}

func (a *App) returnToMenu() {
	if a.gfx != nil {
		if a.gfx.audioPlayer.IsPlaying() {
			a.gfx.audioPlayer.Pause()
		}
		a.gfx.audioPlayer.Close()
		a.gfx = nil
	}
	a.menu.Reset()
	ebiten.SetWindowSize(MenuW*MenuScale, MenuH*MenuScale)
	ebiten.SetWindowTitle("CHIP-8 Emulator")
	a.state = AppStateMenu
}

func (a *App) Update() error {
	switch a.state {
	case AppStateMenu:
		a.menu.Update()
		if a.menu.IsSelected() {
			if err := a.launchGame(a.menu.selectedGame); err != nil {
				return err
			}
		}
	case AppStatePlaying:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.returnToMenu()
			return nil
		}
		return a.gfx.Update()
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	switch a.state {
	case AppStateMenu:
		a.menu.Draw(screen)
	case AppStatePlaying:
		a.gfx.Draw(screen)
	}
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	switch a.state {
	case AppStateMenu:
		return MenuW, MenuH
	case AppStatePlaying:
		return a.gfx.Layout(outsideWidth, outsideHeight)
	}
	return MenuW, MenuH
}

func main() {
	games, err := loadGames()
	if err != nil {
		panic(err)
	}

	var romName string
	flag.StringVar(&romName, "rom", "", "ROM file to load diretamente (nome da pasta em games/)")
	flag.Parse()

	audioCtx := audio.NewContext(sampleRate)

	app := &App{
		audioCtx: audioCtx,
		menu:     NewMenu(games),
		state:    AppStateMenu,
	}

	ebiten.SetWindowSize(MenuW*MenuScale, MenuH*MenuScale)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	// Atalho via flag -rom pula o menu e inicia o jogo diretamente
	if romName != "" {
		game, ok := games[romName]
		if !ok {
			panic("ROM não encontrada: " + romName)
		}
		if err := app.launchGame(&game); err != nil {
			panic(err)
		}
	}

	if err := ebiten.RunGame(app); err != nil {
		panic(err)
	}
}
