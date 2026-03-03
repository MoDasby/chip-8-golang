package main

import "math/rand"

func process(instruction uint16) {
	category := instruction >> 12

	switch category {
	case 0x1: // 1NNN (Jump): Pula para o endereço NNN
		pc = instruction & 0x0FFF
	case 0x2: // 2NNN cria uma subrotina no endereço NNN
		stack[sp] = pc
		pc = instruction & 0x0FFF
		sp++
	case 0x0:
		op := instruction & 0x00FF

		switch op {
		case 0xEE: // 00EE retorna de uma subrotina
			sp--
			pc = stack[sp]
		}

	case 0xA: // Define o registrador I para o endereço NNN
		i = instruction & 0x0FFF
	case 0x6: // 6XNN Coloca o valor NN dentro do registrador VX
		index := (instruction & 0x0F00) >> 8

		v[index] = byte(instruction & 0x00FF)
	case 0x7: // 7XNN (Add VX): Soma o valor NN ao registrador VX
		index := (instruction & 0x0F00) >> 8

		v[index] = v[index] + byte(instruction&0x00FF)
	case 0x3: // 3XNN (Skip if VX == NN): * Execução: Se V[X] for igual a NN, pule uma instrução
		index := (instruction & 0x0F00) >> 8

		if v[index] == byte(instruction&0x00FF) {
			pc += 2
		}
	case 0x4: // 4XNN (Skip if VX != NN): * Execução: Se V[X] for diferente de NN, pule uma instrução
		index := (instruction & 0x0F00) >> 8

		if v[index] != byte(instruction&0x00FF) {
			pc += 2
		}
	case 0x5: // 5XYN (Skip if VX == NN): * Execução: Se V[X] for igual a V[Y], pule uma instrução
		indexX := (instruction & 0x0F00) >> 8
		indexY := (instruction & 0x00F0) >> 4

		if v[indexX] == v[indexY] {
			pc += 2
		}
	case 0x9: // 9XY0: Pula uma instrução se V[X] != V[Y]
		indexX := (instruction & 0xF00) >> 8
		indexY := (instruction & 0xF0) >> 4

		if v[indexX] != v[indexY] {
			pc += 2
		}
	case 0xC: // CXNN: V[X] recebe um número aleatório (0 a 255) AND NN
		index := (instruction & 0x0F00) >> 8
		nn := byte(instruction & 0x00FF)

		randomNumber := byte(rand.Intn(256))
		v[index] = randomNumber & nn
	case 0xB: // BNNN: Pula para o endereço NNN + V[0]
		nnn := instruction & 0x0FFF
		pc = nnn + uint16(v[0])
	case 0xF:
		index := (instruction & 0x0F00) >> 8

		op := instruction & 0x00FF

		switch op {
		case 0x7: // FX07 Copia p valor atual do temporizador de atraso para V[X]
			v[index] = delayTimer
		case 0x15: // FX15 O temporizador de atraso recebe o valor que está em V[X]
			delayTimer = v[index]
		case 0x18: // FX18 O temporizador de som recebe o valor de V[X]
			soundTimer = v[index]
		case 0x1E: // FX1E Soma o valor de V[X] ao registador de índice I
			i += uint16(v[index])
		case 0x55: // FX55 pega em todos os valores dos registadores desde V[0] até V[X], e guarda-os sequencialmente na memória, começando no endereço apontado por I
			for vi := uint16(0); vi <= index; vi++ {
				memory[i+vi] = v[vi]
			}
		case 0x65: // FX65 Lê os valores da memória (começando no endereço I) e preenche os registadores desde V[0] até V[X]
			for vi := uint16(0); vi <= index; vi++ {
				v[vi] = memory[i+vi]
			}
		case 0x29: // FX29 Apontar I para um caractere de texto que está armazenado em v[X]
			i = 0x50 + uint16(v[index]*5)
		case 0x33: // FX33 Guarda um valor decimal na memória
			num := v[index]

			memory[i] = num / 100
			memory[i+1] = (num / 10) % 10
			memory[i+2] = num % 10
		case 0x0A: // FX0A pausa a execução até que qualquer tecla seja pressionada, quando isso acontecer guarda em v[X] a tecla que foi pressionada
			anyPressed := -1

			for i, pressed := range keyboard {
				if pressed {
					anyPressed = i
					break
				}
			}

			if anyPressed < 0 {
				pc -= 2
			} else {
				v[index] = byte(anyPressed)
			}
		}

	case 0x8:
		op := instruction & 0x000F

		indexX := (instruction & 0x0F00) >> 8
		indexY := (instruction & 0x00F0) >> 4

		switch op {
		case 0x0: // 8XY0 copia o valor de V[Y] para V[X]
			v[indexX] = v[indexY]
		case 0x1: // 8XY1 Faz a operação bit a bit OR. (V[X] = V[X] | V[Y]
			v[indexX] = v[indexX] | v[indexY]
		case 0x2: // 8XY2 Faz a operação bit a bit AND. (V[X] = V[X] & V[Y]
			v[indexX] = v[indexX] & v[indexY]
		case 0x3: // 8XY3 Faz a operação bit a bit Ou Exclusivo. (V[X] = V[X] ^ V[Y]
			v[indexX] = v[indexX] ^ v[indexY]
		case 0x4: // 8XY4 Soma V[X] com V[Y].
			sum := uint16(v[indexX]) + uint16(v[indexY])

			carry := sum > 255

			var flag byte = 0

			if carry {
				v[indexX] = byte(sum & 0xFF)
				flag = 1
			} else {
				v[indexX] = byte(sum)
				flag = 0
			}

			v[0xF] = flag
		case 0x5: // 8XY5 Subtrai V[Y] de V[X]
			var flag byte = 0

			if v[indexX] >= v[indexY] {
				flag = 1
			}

			if v[indexX] < v[indexY] {
				flag = 0
			}

			v[indexX] -= v[indexY]

			v[0xF] = flag
		case 0x7: // 8XY7 Subtrai V[X] de V[Y]
			var flag byte = 0

			if v[indexY] >= v[indexX] {
				flag = 1
			}

			if v[indexY] < v[indexX] {
				flag = 0
			}

			v[indexX] = v[indexY] - v[indexX]
			v[0xF] = flag
		case 0x6: // 8XY6 Empurra os bits de V[X] uma casa para a direita (V[X] = V[X] >> 1).
			flag := v[indexX] & 0x1

			v[indexX] >>= 1
			v[0xF] = flag
		case 0xE: // 8XYE Empurra os bits de V[X] uma casa para a esquerda (V[X] = V[X] >> 1).
			flag := (v[indexX] & 0x80) >> 7

			v[indexX] <<= 1
			v[0xF] = flag
		}
	case 0xE:
		op := instruction & 0xFF

		index := (instruction & 0xF00) >> 8

		switch op {
		case 0x9E: // EX9E se o indice v[X] do array de teclado for verdadeiro, pula uma instrução
			if keyboard[v[index]] {
				pc += 2
			}
		case 0xA1: // EXA1 se o indice v[X] do array de teclado for falso, pula uma instrução
			if !keyboard[v[index]] {
				pc += 2
			}
		}
	case 0xD: // DXYN desenha na tela
		indexX := (instruction & 0xF00) >> 8
		indexY := (instruction & 0xF0) >> 4
		n := instruction & 0xF

		x := v[indexX] % 64
		y := v[indexY] % 32

		v[0xF] = 0

		for line := range n {
			sprite := memory[i+line]

			for column := range 8 {
				pixel := (sprite >> (7 - column)) & 1

				finalX := x + byte(column)
				finalY := y + byte(line)

				if finalX > 63 || finalY > 31 {
					continue
				}

				if pixel == 1 {

					if screen[finalY][finalX] == 1 {
						v[0xF] = 1
					}

					screen[finalY][finalX] ^= 1
				}
			}
		}
		drawFlag = true
	}
}
