package main

var memory []byte = make([]byte, 4096) // memória RAM de 4KB

var pc uint16 = 0x200 // Program counter, que aponta para a próxima instrução, começa em 0x200 porque os primeiros 512 bytes (0x000 a 0x1FF) são reservados para o sistema e fontes de caracteres
var i uint16 = 0      // registro de índice, usado para armazenar endereços de memória
var v [16]byte        // registradores V0 a VF, onde VF é usado como flag para algumas operações
var stack [16]uint16  // pilha para armazenar os endereços de retorno quando sub-rotinas são chamadas
var sp uint16 = 0     // stack pointer, que aponta para o topo da pilha

var delayTimer byte // timer de atraso, que é decrementado a uma taxa de 60Hz quando for maior que zero
var soundTimer byte // timer de som, que é decrementado a uma taxa de 60Hz quando for maior que zero. Quando chega a zero, um som é emitido

var keyboard [16]bool   // array com o estado de cada tecla do chip-8
var screen [32][64]byte // mapeamento de pixels da tela
var drawFlag = false    // controle se deve ou não redenhar a tela
