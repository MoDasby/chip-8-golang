package main

import (
	"math"
)

const sampleRate = 44100

type beepStream struct {
	pos int
}

func (s *beepStream) Read(b []byte) (int, error) {
	freq := 440.0        // Frequência do beep (440Hz)
	length := len(b) / 4 // 4 bytes por sample (2 canais de 16-bits)

	for i := 0; i < length; i++ {
		// Calcula a curva da onda
		val := math.Sin(2.0 * math.Pi * freq * float64(s.pos+i) / sampleRate)

		// Converte para int16 controlando o volume (10000 é um bom volume médio)
		v := int16(val * 10000)

		// Preenche Canal Esquerdo
		b[4*i] = byte(v)
		b[4*i+1] = byte(v >> 8)
		// Preenche Canal Direito
		b[4*i+2] = byte(v)
		b[4*i+3] = byte(v >> 8)
	}
	s.pos += length
	return len(b), nil
}
