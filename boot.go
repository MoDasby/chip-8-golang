package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

