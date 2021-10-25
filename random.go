package txfuzz

import (
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
)

func randomHash() common.Hash {
	b := make([]byte, 32)
	rand.Read(b)
	return common.BytesToHash(b)
}

func randomAddress() common.Address {
	switch rand.Int31n(5) {
	case 0, 1, 2:
		b := make([]byte, 20)
		rand.Read(b)
		return common.BytesToAddress(b)
	case 3:
		return common.Address{}
	case 4:
		return common.HexToAddress(ADDR)
	}
	return common.Address{}
}
