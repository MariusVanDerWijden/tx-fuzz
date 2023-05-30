package txfuzz

import (
	"crypto/rand"
	"fmt"
	mathRand "math/rand"

	"github.com/ethereum/go-ethereum/common"
)

const (
	maxDataPerTx = 1 << 17 // 128Kb
)

func randomHash() common.Hash {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return common.BytesToHash(b)
}

func randomAddress() common.Address {
	switch mathRand.Int31n(5) {
	case 0, 1, 2:
		b := make([]byte, 20)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}
		return common.BytesToAddress(b)
	case 3:
		return common.Address{}
	case 4:
		return common.HexToAddress(ADDR)
	}
	return common.Address{}
}

func randomBlobData() ([]byte, error) {
	size := mathRand.Intn(maxDataPerTx)
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
