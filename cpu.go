package main

import (
	"math/rand"
	"os"
)

type CPU struct {
	memory          []byte       // memória RAM de 4KB
	pc              uint16       // Program counter, que aponta para a próxima instrução, começa em 0x200 porque os primeiros 512 bytes (0x000 a 0x1FF) são reservados para o sistema e fontes de caracteres
	i               uint16       // registro de índice, usado para armazenar endereços de memória
	v               []byte       // registradores V0 a VF, onde VF é usado como flag para algumas operações
	stack           []uint16     // pilha para armazenar os endereços de retorno quando sub-rotinas são chamadas
	sp              uint16       // stack pointer, que aponta para o topo da pilha
	delayTimer      byte         // timer de atraso, que é decrementado a uma taxa de 60Hz quando for maior que zero
	soundTimer      byte         // timer de som, que é decrementado a uma taxa de 60Hz quando for maior que zero. Quando chega a zero, um som é emitido
	keyboard        []bool       // array com o estado de cada tecla do chip-8(16 teclas)
	screen          [32][64]byte // mapeamento de pixels da tela
	clearScreenFlag bool         // controle se deve ou não limpar a tela
}

func NewCpu() *CPU {
	return &CPU{
		memory:          make([]byte, 4096),
		pc:              0x200,
		i:               0,
		v:               make([]byte, 16),
		stack:           make([]uint16, 16),
		sp:              0,
		delayTimer:      0,
		soundTimer:      0,
		keyboard:        make([]bool, 16),
		screen:          [32][64]byte{},
		clearScreenFlag: false,
	}
}

func (c *CPU) LoadFontset() {
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

	copy(c.memory[80:], fontset)
}

func (c *CPU) LoadROM(filename string) error {
	rom, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	copy(c.memory[0x200:], rom)

	return nil
}

func (c *CPU) updateKeyboard(i uint16, state bool) {
	c.keyboard[i] = state
}

func (c *CPU) fetch() {

	if c.clearScreenFlag {
		for y := range 32 {
			for x := range 64 {
				c.screen[y][x] = 0
			}
		}

		c.clearScreenFlag = false
	}

	if c.delayTimer > 0 {
		c.delayTimer--
	}
	if c.soundTimer > 0 {
		c.soundTimer--
	}

	cyclesPerFrame := 12
	for i := 0; i < cyclesPerFrame; i++ {
		instruction := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
		c.pc += 2

		c.process(instruction)
	}
}

func (c *CPU) process(instruction uint16) {
	if instruction == 0x00E0 {
		c.clearScreenFlag = true

		return
	}

	category := instruction >> 12

	switch category {
	case 0x1: // 1NNN (Jump): Pula para o endereço NNN
		c.pc = instruction & 0x0FFF
	case 0x2: // 2NNN cria uma subrotina no endereço NNN
		c.stack[c.sp] = c.pc
		c.pc = instruction & 0x0FFF
		c.sp++
	case 0x0:
		op := instruction & 0x00FF

		switch op {
		case 0xEE: // 00EE retorna de uma subrotina
			c.sp--
			c.pc = c.stack[c.sp]
		}

	case 0xA: // Define o registrador I para o endereço NNN
		c.i = instruction & 0x0FFF
	case 0x6: // 6XNN Coloca o valor NN dentro do registrador VX
		index := (instruction & 0x0F00) >> 8

		c.v[index] = byte(instruction & 0x00FF)
	case 0x7: // 7XNN (Add VX): Soma o valor NN ao registrador VX
		index := (instruction & 0x0F00) >> 8

		c.v[index] = c.v[index] + byte(instruction&0x00FF)
	case 0x3: // 3XNN (Skip if VX == NN): * Execução: Se V[X] for igual a NN, pule uma instrução
		index := (instruction & 0x0F00) >> 8

		if c.v[index] == byte(instruction&0x00FF) {
			c.pc += 2
		}
	case 0x4: // 4XNN (Skip if VX != NN): * Execução: Se V[X] for diferente de NN, pule uma instrução
		index := (instruction & 0x0F00) >> 8

		if c.v[index] != byte(instruction&0x00FF) {
			c.pc += 2
		}
	case 0x5: // 5XYN (Skip if VX == NN): * Execução: Se V[X] for igual a V[Y], pule uma instrução
		indexX := (instruction & 0x0F00) >> 8
		indexY := (instruction & 0x00F0) >> 4

		if c.v[indexX] == c.v[indexY] {
			c.pc += 2
		}
	case 0x9: // 9XY0: Pula uma instrução se V[X] != V[Y]
		indexX := (instruction & 0xF00) >> 8
		indexY := (instruction & 0xF0) >> 4

		if c.v[indexX] != c.v[indexY] {
			c.pc += 2
		}
	case 0xC: // CXNN: V[X] recebe um número aleatório (0 a 255) AND NN
		index := (instruction & 0x0F00) >> 8
		nn := byte(instruction & 0x00FF)

		randomNumber := byte(rand.Intn(256))
		c.v[index] = randomNumber & nn
	case 0xB: // BNNN: Pula para o endereço NNN + V[0]
		nnn := instruction & 0x0FFF
		c.pc = nnn + uint16(c.v[0])
	case 0xF:
		index := (instruction & 0x0F00) >> 8

		op := instruction & 0x00FF

		switch op {
		case 0x7: // FX07 Copia p valor atual do temporizador de atraso para V[X]
			c.v[index] = c.delayTimer
		case 0x15: // FX15 O temporizador de atraso recebe o valor que está em V[X]
			c.delayTimer = c.v[index]
		case 0x18: // FX18 O temporizador de som recebe o valor de V[X]
			c.soundTimer = c.v[index]
		case 0x1E: // FX1E Soma o valor de V[X] ao registador de índice I
			c.i += uint16(c.v[index])
		case 0x55: // FX55 pega em todos os valores dos registadores desde V[0] até V[X], e guarda-os sequencialmente na memória, começando no endereço apontado por I
			for vi := uint16(0); vi <= index; vi++ {
				c.memory[c.i+vi] = c.v[vi]
			}
		case 0x65: // FX65 Lê os valores da memória (começando no endereço I) e preenche os registadores desde V[0] até V[X]
			for vi := uint16(0); vi <= index; vi++ {
				c.v[vi] = c.memory[c.i+vi]
			}
		case 0x29: // FX29 Apontar I para um caractere de texto que está armazenado em v[X]
			c.i = 0x50 + uint16(c.v[index]*5)
		case 0x33: // FX33 Guarda um valor decimal na memória
			num := c.v[index]

			c.memory[c.i] = num / 100
			c.memory[c.i+1] = (num / 10) % 10
			c.memory[c.i+2] = num % 10
		case 0x0A: // FX0A pausa a execução até que qualquer tecla seja pressionada, quando isso acontecer guarda em v[X] a tecla que foi pressionada
			anyPressed := -1

			for key, pressed := range c.keyboard {
				if pressed {
					anyPressed = key
					break
				}
			}

			if anyPressed < 0 {
				c.pc -= 2
			} else {
				c.v[index] = byte(anyPressed)
			}
		}

	case 0x8:
		op := instruction & 0x000F

		indexX := (instruction & 0x0F00) >> 8
		indexY := (instruction & 0x00F0) >> 4

		switch op {
		case 0x0: // 8XY0 copia o valor de V[Y] para V[X]
			c.v[indexX] = c.v[indexY]
		case 0x1: // 8XY1 Faz a operação bit a bit OR. (V[X] = V[X] | V[Y]
			c.v[indexX] = c.v[indexX] | c.v[indexY]
		case 0x2: // 8XY2 Faz a operação bit a bit AND. (V[X] = V[X] & V[Y]
			c.v[indexX] = c.v[indexX] & c.v[indexY]
		case 0x3: // 8XY3 Faz a operação bit a bit Ou Exclusivo. (V[X] = V[X] ^ V[Y]
			c.v[indexX] = c.v[indexX] ^ c.v[indexY]
		case 0x4: // 8XY4 Soma V[X] com V[Y].
			sum := uint16(c.v[indexX]) + uint16(c.v[indexY])

			carry := sum > 255

			var flag byte = 0

			if carry {
				c.v[indexX] = byte(sum & 0xFF)
				flag = 1
			} else {
				c.v[indexX] = byte(sum)
				flag = 0
			}

			c.v[0xF] = flag
		case 0x5: // 8XY5 Subtrai V[Y] de V[X]
			var flag byte = 0

			if c.v[indexX] >= c.v[indexY] {
				flag = 1
			}

			if c.v[indexX] < c.v[indexY] {
				flag = 0
			}

			c.v[indexX] -= c.v[indexY]

			c.v[0xF] = flag
		case 0x7: // 8XY7 Subtrai V[X] de V[Y]
			var flag byte = 0

			if c.v[indexY] >= c.v[indexX] {
				flag = 1
			}

			if c.v[indexY] < c.v[indexX] {
				flag = 0
			}

			c.v[indexX] = c.v[indexY] - c.v[indexX]
			c.v[0xF] = flag
		case 0x6: // 8XY6 Empurra os bits de V[X] uma casa para a direita (V[X] = V[X] >> 1).
			flag := c.v[indexX] & 0x1

			c.v[indexX] >>= 1
			c.v[0xF] = flag
		case 0xE: // 8XYE Empurra os bits de V[X] uma casa para a esquerda (V[X] = V[X] >> 1).
			flag := (c.v[indexX] & 0x80) >> 7

			c.v[indexX] <<= 1
			c.v[0xF] = flag
		}
	case 0xE:
		op := instruction & 0xFF

		index := (instruction & 0xF00) >> 8

		switch op {
		case 0x9E: // EX9E se o indice v[X] do array de teclado for verdadeiro, pula uma instrução
			if c.keyboard[c.v[index]] {
				c.pc += 2
			}
		case 0xA1: // EXA1 se o indice v[X] do array de teclado for falso, pula uma instrução
			if !c.keyboard[c.v[index]] {
				c.pc += 2
			}
		}
	case 0xD: // DXYN desenha na tela
		indexX := (instruction & 0xF00) >> 8
		indexY := (instruction & 0xF0) >> 4
		n := instruction & 0xF

		x := c.v[indexX] % 64
		y := c.v[indexY] % 32

		c.v[0xF] = 0

		for line := uint16(0); line < n; line++ {
			sprite := c.memory[c.i+line]

			for column := byte(0); column < 8; column++ {
				pixel := (sprite >> (7 - column)) & 1

				finalX := x + column
				finalY := y + byte(line)

				if finalX > 63 || finalY > 31 {
					continue
				}

				if pixel == 1 {

					if c.screen[finalY][finalX] == 1 {
						c.v[0xF] = 1
					}

					c.screen[finalY][finalX] ^= 1
				}
			}
		}
	}
}
