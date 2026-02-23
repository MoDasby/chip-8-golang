package audio

import (
	"encoding/binary"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

type Beeper struct {
	Dev        sdl.AudioDeviceID
	Spec       sdl.AudioSpec
	buf        []byte
	playing    bool
	targetQ    uint32 // quanto manter na fila (bytes)
	sampleRate float64
	phase      float64
	freq       float64
	amp        float64
}

// InitBeeper abre um device "non-callback" (Callback=nil) e prepara um buffer de tom.
// Chame sdl.Init(sdl.INIT_AUDIO) antes, em algum lugar do seu programa.
func InitBeeper(freqHz float64, volume01 float64) (*Beeper, error) {
	desired := &sdl.AudioSpec{
		Freq:     44100,
		Format:   sdl.AUDIO_S16SYS,
		Channels: 1,
		Samples:  1024,
		Callback: nil, // <- importante: device sem callback
	}

	var obtained sdl.AudioSpec
	dev, err := sdl.OpenAudioDevice("", false, desired, &obtained, 0)
	if err != nil {
		return nil, err
	}

	b := &Beeper{
		Dev:        dev,
		Spec:       obtained,
		sampleRate: float64(obtained.Freq),
		freq:       freqHz,
		amp:        clamp(volume01, 0, 1) * 0.25, // 0.25 pra não estourar fácil
		// manter ~100ms na fila por padrão:
		targetQ: uint32(obtained.Freq/10) * 2, // int16 mono => 2 bytes por sample
	}

	// gera um chunk curto (~20ms) e reusa
	b.buf = make([]byte, int(float64(obtained.Freq)*0.02)*2)

	// começa o device (despausado); silêncio enquanto não estiver "playing"
	sdl.PauseAudioDevice(dev, false)
	return b, nil
}

// Start liga o beep (começa a enfileirar tom).
func (b *Beeper) Start() { b.playing = true }

// Stop desliga e limpa a fila (corta o som na hora).
func (b *Beeper) Stop() {
	b.playing = false
	sdl.ClearQueuedAudio(b.Dev)
}

// Pump deve ser chamado frequentemente (ex: a cada frame).
// Ele mantém a fila abastecida enquanto estiver tocando.
func (b *Beeper) Pump() error {
	if !b.playing {
		return nil
	}

	queued := sdl.GetQueuedAudioSize(b.Dev)
	for queued < b.targetQ {
		b.fillSine(b.buf) // ou quadrada, etc.
		if err := sdl.QueueAudio(b.Dev, b.buf); err != nil {
			return err
		}
		queued += uint32(len(b.buf))
	}
	return nil
}

// Close fecha o device.
func (b *Beeper) Close() {
	if b.Dev != 0 {
		sdl.CloseAudioDevice(b.Dev)
		b.Dev = 0
	}
}

func (b *Beeper) fillSine(out []byte) {
	// out é PCM S16 mono (little/big depende do AUDIO_S16SYS; aqui usamos binary.LittleEndian
	// porque no Linux x86_64 normalmente é little-endian. Se quiser 100% certo em qualquer CPU,
	// você pode checar endianness ou usar AUDIO_S16LSB explicitamente.
	n := len(out) / 2
	for i := 0; i < n; i++ {
		v := math.Sin(b.phase) * (b.amp * 32767.0)
		s := int16(v)

		binary.LittleEndian.PutUint16(out[i*2:], uint16(s))

		b.phase += 2 * math.Pi * b.freq / b.sampleRate
		if b.phase >= 2*math.Pi {
			b.phase -= 2 * math.Pi
		}
	}
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
