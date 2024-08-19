package txfuzz

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	mathRand "math/rand"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

func randomAuthEntry(f *filler.Filler) *types.Authorization {
	return &types.Authorization{
		ChainID: f.MemInt(),
		Address: randomAddress(),
		Nonce:   []uint64{f.Uint64()},
	}
}

func RandomAuthList(f *filler.Filler, sk *ecdsa.PrivateKey) (types.AuthorizationList, error) {
	var authList types.AuthorizationList
	entries := f.MemInt()
	for i := 0; i < int(entries.Uint64()); i++ {
		signed, err := types.SignAuth(randomAuthEntry(f), sk)
		if err != nil {
			return nil, err
		}
		authList = append(authList, signed)
	}
	return authList, nil
}
