package main

import (
	"embed"
	"encoding/json"
	"path"
)

const GAMES_FOLDER_NAME = "games"

//go:embed games
var GAMES_FOLDER embed.FS

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
	gamesFolder, err := GAMES_FOLDER.ReadDir(GAMES_FOLDER_NAME)
	if err != nil {
		return nil, err
	}

	games := make(map[string]GameMetadata)

	for _, entry := range gamesFolder {
		if entry.IsDir() {
			file, err := GAMES_FOLDER.ReadFile(path.Join(GAMES_FOLDER_NAME, entry.Name(), "metadata.json"))
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

func readROM(metadata *GameMetadata) ([]byte, error) {
	romPath := path.Join(GAMES_FOLDER_NAME, metadata.FolderName, metadata.RomLocation)
	return GAMES_FOLDER.ReadFile(romPath)
}
