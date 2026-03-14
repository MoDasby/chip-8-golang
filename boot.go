package main

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sort"

	"github.com/manifoldco/promptui"
)

const (
	GAMES_FOLDER = "games"
)

type GameMetadata struct {
	Name        string `json:"name"`
	FolderName  string `json:"folderName"`
	Description string `json:"description"`
	RomLocation string `json:"romLocation"`
	Theme       struct {
		PrimaryColor   string `json:"primaryColor"`
		SecondaryColor string `json:"secondaryColor"`
	} `json:"theme"`
	Order int `json:"order"`
}

func boot() (*GameMetadata, error) {
	loadFontset()

	games, err := loadGames()
	if err != nil {
		return nil, err
	}

	var romName string
	flag.StringVar(&romName, "rom", "", "ROM file to load")

	flag.Parse()

	if romName != "" {
		game, ok := games[romName]
		if !ok {
			return nil, errors.New("Choosed ROM does not exist")
		}

		loadROM(normalizeRomLocation(&game))

		return &game, nil
	}

	game, err := askForGame(games)
	if err != nil {
		return nil, err
	}

	loadROM(normalizeRomLocation(game))

	return game, nil
}

func loadGames() (map[string]GameMetadata, error) {
	gamesFolder, err := os.ReadDir(GAMES_FOLDER)
	if err != nil {
		return nil, err
	}

	games := make(map[string]GameMetadata)

	for _, entry := range gamesFolder {
		if entry.IsDir() {
			file, err := os.ReadFile(filepath.Join(GAMES_FOLDER, entry.Name(), "metadata.json"))
			if err != nil {
				return nil, err
			}

			var metadata GameMetadata

			if err := json.Unmarshal(file, &metadata); err != nil {
				return nil, err
			}

			games[entry.Name()] = metadata
		}
	}

	return games, nil
}

func normalizeRomLocation(metadata *GameMetadata) string {
	return filepath.Join(GAMES_FOLDER, metadata.FolderName, metadata.RomLocation)
}

func askForGame(gamesMap map[string]GameMetadata) (*GameMetadata, error) {
	var gamesSlice []GameMetadata

	for _, game := range gamesMap {
		gamesSlice = append(gamesSlice, game)
	}

	sort.Slice(gamesSlice, func(i, j int) bool {
		if gamesSlice[i].Order != gamesSlice[j].Order {
			return gamesSlice[i].Order > gamesSlice[j].Order
		}
		return gamesSlice[i].Name < gamesSlice[j].Name
	})

	prompt := promptui.Select{
		Label: "Escolha uma ROM",
		Items: gamesSlice,
		Templates: &promptui.SelectTemplates{
			Active:   "▸ {{ .Name }} - {{ .Description }}",
			Inactive: "  {{ .Name }} - {{ .Description }}",
			Selected: "✔ {{ .Name }} - {{ .Description }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &gamesSlice[index], nil
}

func loadROM(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	copy(memory[0x200:], data)
}

func loadFontset() {
	fontset := []byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80} // F

	copy(memory[80:], fontset)
}
