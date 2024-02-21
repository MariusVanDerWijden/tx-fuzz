// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mutator

import (
	"encoding/binary"
	"math/rand"
	"unsafe"
)

var (
	interesting8  = []int8{-128, -1, 0, 1, 16, 32, 64, 100, 127}
	interesting16 = []int16{-32768, -129, 128, 255, 256, 512, 1000, 1024, 4096, 32767}
	interesting32 = []int32{-2147483648, -100663046, -32769, 32768, 65535, 65536, 100663045, 2147483647}
)

func init() {
	for _, v := range interesting8 {
		interesting16 = append(interesting16, int16(v))
	}
	for _, v := range interesting16 {
		interesting32 = append(interesting32, int32(v))
	}
}

type Mutator struct {
	r *rand.Rand
}

func NewMutator(r *rand.Rand) *Mutator {
	return &Mutator{r: r}
}

func (m *Mutator) rand(n int) int {
	if n <= 0 {
		return 0
	}
	return m.r.Intn(n)
}

func (m *Mutator) bool() bool {
	return m.r.Int()%2 == 0
}

func (m *Mutator) randByteOrder() binary.ByteOrder {
	if m.bool() {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

// chooseLen chooses length of range mutation in range [1,n]. It gives
// preference to shorter ranges.
func (m *Mutator) chooseLen(n int) int {
	switch x := m.rand(100); {
	case x < 90:
		return m.rand(min(8, n)) + 1
	case x < 99:
		return m.rand(min(32, n)) + 1
	default:
		return m.rand(n) + 1
	}
}

func (m *Mutator) FillBytes(ptr *[]byte) {
	m.r.Read(*ptr)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type byteSliceMutator func(*Mutator, []byte) []byte

var byteSliceMutators = []byteSliceMutator{
	byteSliceRemoveBytes,
	byteSliceInsertRandomBytes,
	byteSliceDuplicateBytes,
	byteSliceOverwriteBytes,
	byteSliceBitFlip,
	byteSliceXORByte,
	byteSliceSwapByte,
	byteSliceArithmeticUint8,
	byteSliceArithmeticUint16,
	byteSliceArithmeticUint32,
	byteSliceArithmeticUint64,
	byteSliceOverwriteInterestingUint8,
	byteSliceOverwriteInterestingUint16,
	byteSliceOverwriteInterestingUint32,
	byteSliceInsertConstantBytes,
	byteSliceOverwriteConstantBytes,
	byteSliceShuffleBytes,
	byteSliceSwapBytes,
}

func (m *Mutator) MutateBytes(ptrB *[]byte) {
	b := *ptrB
	defer func() {
		oldHdr := unsafe.SliceData(*ptrB)
		newHdr := unsafe.SliceData(b)
		if oldHdr != newHdr {
			panic("data moved to new address")
		}
		*ptrB = b
	}()

	for {
		mut := byteSliceMutators[m.rand(len(byteSliceMutators))]
		if mutated := mut(m, b); mutated != nil {
			b = mutated
			return
		}
	}
}
