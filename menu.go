package main

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	MenuW          = 640
	MenuH          = 320
	MenuScale      = 2
	listItemHeight = 20
	maxVisible     = 10
)

var (
	colorMenuBg       = color.RGBA{R: 10, G: 10, B: 30, A: 255}
	colorMenuBorder   = color.RGBA{R: 50, G: 50, B: 130, A: 255}
	colorMenuDivider  = color.RGBA{R: 80, G: 80, B: 180, A: 255}
	colorMenuSelected = color.RGBA{R: 30, G: 50, B: 130, A: 255}
)

type Menu struct {
	games        []GameMetadata
	selected     int
	scrollOffset int
	selectedGame *GameMetadata
}

func NewMenu(gamesMap map[string]GameMetadata) *Menu {
	var games []GameMetadata
	for _, g := range gamesMap {
		games = append(games, g)
	}
	sort.Slice(games, func(i, j int) bool {
		if games[i].Order != games[j].Order {
			return games[i].Order > games[j].Order
		}
		return games[i].Name < games[j].Name
	})
	return &Menu{games: games}
}

func (m *Menu) Update() {
	if len(m.games) == 0 {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selected = (m.selected - 1 + len(m.games)) % len(m.games)
		m.adjustScroll()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selected = (m.selected + 1) % len(m.games)
		m.adjustScroll()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g := m.games[m.selected]
		m.selectedGame = &g
	}
}

func (m *Menu) adjustScroll() {
	if m.selected < m.scrollOffset {
		m.scrollOffset = m.selected
	} else if m.selected >= m.scrollOffset+maxVisible {
		m.scrollOffset = m.selected - maxVisible + 1
	}
}

func (m *Menu) IsSelected() bool {
	return m.selectedGame != nil
}

func (m *Menu) Reset() {
	m.selectedGame = nil
}

func (m *Menu) Draw(screen *ebiten.Image) {
	screen.Fill(colorMenuBg)

	// Borda externa
	vector.FillRect(screen, 5, 5, float32(MenuW-10), float32(MenuH-10), colorMenuBorder, false)
	vector.FillRect(screen, 7, 7, float32(MenuW-14), float32(MenuH-14), colorMenuBg, false)

	// Título
	title := "=== COSMAC VIP ==="
	ebitenutil.DebugPrintAt(screen, title, (MenuW-len(title)*6)/2, 14)

	subtitle := "Selecione um jogo"
	ebitenutil.DebugPrintAt(screen, subtitle, (MenuW-len(subtitle)*6)/2, 27)

	// Divisor superior
	vector.FillRect(screen, 15, 42, float32(MenuW-30), 1, colorMenuDivider, false)

	// Lista de jogos
	listStartY := 48
	if len(m.games) == 0 {
		ebitenutil.DebugPrintAt(screen, "Nenhuma ROM encontrada na pasta 'games/'.", 20, listStartY+10)
	}
	for i := 0; i < maxVisible; i++ {
		idx := m.scrollOffset + i
		if idx >= len(m.games) {
			break
		}
		game := m.games[idx]
		y := listStartY + i*listItemHeight

		if idx == m.selected {
			vector.FillRect(screen, 15, float32(y-2), float32(MenuW-30), float32(listItemHeight-1), colorMenuSelected, false)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf(">> %s", game.Name), 20, y+3)
		} else {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("   %s", game.Name), 20, y+3)
		}
	}

	// Indicadores de scroll
	if m.scrollOffset > 0 {
		ebitenutil.DebugPrintAt(screen, "^", MenuW-25, listStartY+2)
	}
	if m.scrollOffset+maxVisible < len(m.games) {
		ebitenutil.DebugPrintAt(screen, "v", MenuW-25, listStartY+(maxVisible-1)*listItemHeight+2)
	}

	// Divisor inferior
	bottomDividerY := float32(listStartY+maxVisible*listItemHeight) + 5
	vector.FillRect(screen, 15, bottomDividerY, float32(MenuW-30), 1, colorMenuDivider, false)

	// Descrição do jogo selecionado
	if len(m.games) > 0 {
		desc := m.games[m.selected].Description
		if len(desc) > 96 {
			desc = desc[:93] + "..."
		}
		ebitenutil.DebugPrintAt(screen, desc, 15, int(bottomDividerY)+8)
	}

	// Instruções
	instructions := "W/S ou Up/Down: Navegar     Enter: Jogar     N: Próxima musica     Esc: Sair"
	instrX := (MenuW - len(instructions)*6) / 2
	if instrX < 10 {
		instrX = 10
	}
	ebitenutil.DebugPrintAt(screen, instructions, instrX, MenuH-20)
}
