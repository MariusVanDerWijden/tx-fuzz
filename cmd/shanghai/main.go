package main

import (
	"encoding/binary"
	"fmt"

	"github.com/MariusVanDerWijden/tx-fuzz/helper"
)

var (
	maxCodeSize     = 24576
	maxInitCodeSize = 2 * maxCodeSize
)

func main() {
	// Store coinbase
	helper.Execute([]byte{0x41, 0x41, 0x55}, 50000)
	// Call coinbase
	// 5x PUSH0, COINBASE, GAS, CALL
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xf1}, 50000)
	// 5x PUSH0, COINBASE, GAS, CALLCODE
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xf2}, 50000)
	// 5x PUSH0, COINBASE, GAS, DELEGATECALL
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xf4}, 50000)
	// 5x PUSH0, COINBASE, GAS, STATICCALL
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xfA}, 50000)
	// COINBASE, SELFDESTRUCT
	helper.Execute([]byte{0x41, 0xff}, 50000)
	// COINBASE, EXTCODESIZE
	helper.Execute([]byte{0x41, 0x3b}, 50000)
	// 3x PUSH0, COINBASE, EXTCODECOPY
	helper.Execute([]byte{0x5f, 0x5f, 0x5f, 0x41, 0x3C}, 50000)
	// COINBASE, EXTCODEHASH
	helper.Execute([]byte{0x41, 0x3F}, 50000)
	// COINBASE, BALANCE
	helper.Execute([]byte{0x41, 0x31}, 50000)
	// loop push0
	// JUMPDEST, PUSH0, JUMP
	helper.Execute([]byte{0x58, 0x5f, 0x56}, 50000)
	fmt.Println("Limit&MeterInitcode")
	// limit & meter initcode
	sizes := []int{
		maxInitCodeSize - 2,
		maxInitCodeSize - 1,
		maxInitCodeSize,
		maxInitCodeSize + 1,
		maxInitCodeSize + 2,
		maxInitCodeSize * 2,
	}
	// size x JUMPDEST STOP
	for _, size := range sizes {
		initcode := repeatOpcode(size, 0x58)
		helper.Execute(append(initcode, 0x00), 50000)
	}
	// size x STOP STOP
	for _, size := range sizes {
		helper.Execute(repeatOpcode(size, 0x00), 50000)
	}
	// PUSH4 size, PUSH0, PUSH0, CREATE
	for _, size := range sizes {
		initcode := pushSize(size)
		helper.Execute(append(initcode, []byte{0x57, 0x57, 0xF0}...), 50000)
	}
	// PUSH4 size, PUSH0, PUSH0, CREATE2
	for _, size := range sizes {
		initcode := pushSize(size)
		helper.Execute(append(initcode, []byte{0x57, 0x57, 0xF5}...), 50000)
	}
}

// PUSH4 size
func pushSize(size int) []byte {
	code := []byte{63}
	sizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeArr, uint32(size))
	code = append(code, sizeArr...)
	return code
}

func repeatOpcode(size int, opcode byte) []byte {
	initcode := []byte{}
	for i := 0; i < size; i++ {
		initcode = append(initcode, opcode)
	}
	return initcode
}
