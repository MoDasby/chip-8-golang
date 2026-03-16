package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"path"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

type BGManager struct {
	songs      []SongMetadata
	playOrder  []int
	nextSong   int
	randomizer *rand.Rand
}

var ErrUnsupportedAudioFormat = errors.New("formato de áudio não suportado")

func NewBGManager() *BGManager {
	return &BGManager{
		randomizer: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type SongMetadata struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	File   string `json:"file"`
}

type BGMSong struct {
	metadata *SongMetadata
	buffer   io.Reader
}

const SONGS_FOLDER_NAME = "music"

//go:embed music
var SONGS_FOLDER embed.FS

func (bgm *BGManager) loadSongs() error {
	songsFolder, err := SONGS_FOLDER.ReadFile(path.Join(SONGS_FOLDER_NAME, "metadata.json"))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(songsFolder, &bgm.songs); err != nil {
		return err
	}

	bgm.resetPlayOrder()

	fmt.Printf("musicas carregadas: %d\n", len(bgm.songs))

	return nil
}

func (bgm *BGManager) next() (*BGMSong, error) {
	if len(bgm.songs) == 0 {
		return nil, fmt.Errorf("nenhuma música carregada")
	}

	if len(bgm.playOrder) != len(bgm.songs) {
		bgm.resetPlayOrder()
	}

	if bgm.nextSong >= len(bgm.playOrder) {
		bgm.resetPlayOrder()
	}

	metadata := &bgm.songs[bgm.playOrder[bgm.nextSong]]
	bgm.nextSong++

	buf, err := bgm.read(metadata)
	if err != nil {
		return nil, err
	}

	return &BGMSong{
		metadata: metadata,
		buffer:   buf,
	}, nil
}

func (bgm *BGManager) resetPlayOrder() {
	bgm.playOrder = make([]int, len(bgm.songs))
	for i := range bgm.songs {
		bgm.playOrder[i] = i
	}

	if len(bgm.playOrder) > 1 {
		bgm.randomizer.Shuffle(len(bgm.playOrder), func(i, j int) {
			bgm.playOrder[i], bgm.playOrder[j] = bgm.playOrder[j], bgm.playOrder[i]
		})
	}

	bgm.nextSong = 0
}

func (bgm *BGManager) read(metadata *SongMetadata) (io.Reader, error) {
	file, err := SONGS_FOLDER.ReadFile(path.Join(SONGS_FOLDER_NAME, metadata.File))
	if err != nil {
		return nil, err
	}

	stream, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(file))
	if err != nil {
		return nil, err
	}

	return stream, nil
}
