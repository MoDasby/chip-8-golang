package main

import (
	"errors"
	"flag"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

const (
	popupEnterDuration = 280 * time.Millisecond
	popupVisibleTime   = 3200 * time.Millisecond
	popupExitDuration  = 280 * time.Millisecond
	popupMargin        = 12
	popupPadding       = 8
)

var (
	popupBgColor     = color.RGBA{R: 12, G: 12, B: 24, A: 232}
	popupBorderColor = color.RGBA{R: 84, G: 84, B: 180, A: 255}
)

type TrackPopup struct {
	Title     string
	Author    string
	StartedAt time.Time
	Active    bool
}

type App struct {
	state      AppState
	menu       *Menu
	gfx        *GFX
	audioCtx   *audio.Context
	bgm        *BGManager
	bgmPlayer  *audio.Player
	bgmRetryAt time.Time
	trackPopup TrackPopup
}

func (a *App) launchGame(game *GameMetadata) error {
	cpu := NewCpu()
	cpu.LoadFontset()

	rom, err := readROM(game)
	if err != nil {
		return err
	}

	cpu.LoadROM(rom)

	primaryRGB, _ := convertRGB(strings.Split(game.Theme.PrimaryColor, ","))
	secondaryRGB, _ := convertRGB(strings.Split(game.Theme.SecondaryColor, ","))

	beepPlayer, err := a.audioCtx.NewPlayer(&beepStream{})
	if err != nil {
		return err
	}

	a.gfx = &GFX{
		PrimaryColor:   color.RGBA{R: primaryRGB[0], G: primaryRGB[1], B: primaryRGB[2], A: 255},
		SecondaryColor: color.RGBA{R: secondaryRGB[0], G: secondaryRGB[1], B: secondaryRGB[2], A: 255},
		cpu:            cpu,
		audioPlayer:    beepPlayer,
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
		if a.gfx.offscreen != nil {
			a.gfx.offscreen.Deallocate()
			a.gfx.offscreen = nil
		}
		a.gfx = nil
	}
	a.menu.Reset()
	ebiten.SetWindowSize(MenuW*MenuScale, MenuH*MenuScale)
	ebiten.SetWindowTitle("CHIP-8 Emulator")
	a.state = AppStateMenu
}

func (a *App) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyN) && a.bgmPlayer != nil && a.bgmPlayer.IsPlaying() {
		a.bgmPlayer.Close()
		a.bgmPlayer = nil
	}

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
		} else {
			if err := a.gfx.Update(); err != nil {
				return err
			}
		}
	}

	a.updateBGM()
	a.updateTrackPopup()

	return nil
}

func (a *App) updateTrackPopup() {
	if !a.trackPopup.Active {
		return
	}

	total := popupEnterDuration + popupVisibleTime + popupExitDuration
	if time.Since(a.trackPopup.StartedAt) >= total {
		a.trackPopup.Active = false
	}
}

func (a *App) updateBGM() {
	if a.bgm == nil || a.audioCtx == nil {
		return
	}

	now := time.Now()
	if !a.bgmRetryAt.IsZero() && now.Before(a.bgmRetryAt) {
		return
	}

	if a.bgmPlayer != nil && a.bgmPlayer.IsPlaying() {
		return
	}

	if a.bgmPlayer != nil {
		a.bgmPlayer.Close()
		a.bgmPlayer = nil
	}

	next, err := a.bgm.next()
	if err != nil {
		if errors.Is(err, ErrUnsupportedAudioFormat) {
			fmt.Println("BGM desativado:", err)
			a.bgm = nil
			return
		}

		fmt.Println("Falha ao carregar próxima música", err)
		a.bgmRetryAt = now.Add(2 * time.Second)
		return
	}

	player, err := a.audioCtx.NewPlayer(next.buffer)
	if err != nil {
		fmt.Println("Falha ao criar player", err)
		a.bgmRetryAt = now.Add(2 * time.Second)
		return
	}

	a.bgmPlayer = player
	a.bgmRetryAt = time.Time{}
	a.bgmPlayer.Play()
	a.showTrackPopup(next.metadata)
}

func (a *App) showTrackPopup(metadata *SongMetadata) {
	if metadata == nil {
		return
	}

	a.trackPopup = TrackPopup{
		Title:     metadata.Title,
		Author:    metadata.Author,
		StartedAt: time.Now(),
		Active:    true,
	}
}

func (a *App) Draw(screen *ebiten.Image) {
	switch a.state {
	case AppStateMenu:
		a.menu.Draw(screen)
	case AppStatePlaying:
		a.gfx.Draw(screen)
	}

	a.drawTrackPopup(screen)
}

func (a *App) drawTrackPopup(screen *ebiten.Image) {
	if !a.trackPopup.Active {
		return
	}

	elapsed := time.Since(a.trackPopup.StartedAt)
	total := popupEnterDuration + popupVisibleTime + popupExitDuration
	if elapsed >= total {
		a.trackPopup.Active = false
		return
	}

	popUpTitleLine := "Tocando agora"

	titleLine := a.trackPopup.Title
	authorLine := fmt.Sprintf("by %s", a.trackPopup.Author)
	if strings.TrimSpace(a.trackPopup.Author) == "" {
		authorLine = ""
	}

	lineHeight := 12
	lineCount := 2
	if authorLine != "" {
		lineCount++
	}

	line1W := len(popUpTitleLine) * 6
	line2W := len(titleLine) * 6
	line3W := len(authorLine) * 6
	textW := int(max(float64(line1W), float64(line2W), float64(line3W)))
	boxW := textW + popupPadding*2
	boxH := lineCount*lineHeight + popupPadding*2

	scale := 1.0
	if a.state == AppStatePlaying {
		scale = 2.0
	}

	screenW, _ := screen.Bounds().Dx(), screen.Bounds().Dy()
	drawW := float64(boxW) * scale
	xVisible := float64(screenW) - float64(popupMargin)*scale - drawW
	xHidden := float64(screenW) + float64(popupMargin)*scale
	y := float64(popupMargin) * scale

	var x float64
	if elapsed < popupEnterDuration {
		p := float64(elapsed) / float64(popupEnterDuration)
		x = xHidden + (xVisible-xHidden)*p
	} else if elapsed < popupEnterDuration+popupVisibleTime {
		x = xVisible
	} else {
		exitElapsed := elapsed - popupEnterDuration - popupVisibleTime
		p := float64(exitElapsed) / float64(popupExitDuration)
		x = xVisible + (xHidden-xVisible)*p
	}

	popupImage := ebiten.NewImage(boxW, boxH)
	popupImage.Fill(popupBgColor)
	vector.StrokeRect(popupImage, 0, 0, float32(boxW), float32(boxH), 1, popupBorderColor, false)

	tx := popupPadding
	ty := popupPadding

	ebitenutil.DebugPrintAt(popupImage, popUpTitleLine, tx, ty-4)
	ebitenutil.DebugPrintAt(popupImage, titleLine, tx, ty+lineHeight)
	if authorLine != "" {
		ebitenutil.DebugPrintAt(popupImage, authorLine, tx, ty+lineHeight*2)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	screen.DrawImage(popupImage, op)
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

	bgm := NewBGManager()

	if err := bgm.loadSongs(); err != nil {
		panic(err)
	}

	app := &App{
		audioCtx: audioCtx,
		menu:     NewMenu(games),
		state:    AppStateMenu,
		bgm:      bgm,
	}

	ebiten.SetWindowSize(MenuW*MenuScale, MenuH*MenuScale)
	ebiten.SetWindowTitle("CHIP-8")

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
