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

func boot(cpu *CPU) (*GameMetadata, error) {
	cpu.LoadFontset()

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

		if err := cpu.LoadROM(normalizeRomLocation(&game)); err != nil {
			return nil, err
		}

		return &game, nil
	}

	game, err := askForGame(games)
	if err != nil {
		return nil, err
	}

	if err := cpu.LoadROM(normalizeRomLocation(game)); err != nil {
		return nil, err
	}

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
