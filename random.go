package txfuzz

import (
	"fmt"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
)

const (
	maxDataPerTx = 1 << 19 // 512Kb
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

func randomBlobData() ([]byte, error) {
	size := rand.Intn(maxDataPerTx)
	data := make([]byte, size)
	n, err := rand.Read(data)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("could not create random blob data with size %d: %v", size, err)
	}
	return data, nil
}
