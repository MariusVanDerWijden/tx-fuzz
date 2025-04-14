package txfuzz

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	mathRand "math/rand"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
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
	switch mathRand.Int31n(8) {
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
	case 5:
		return params.BeaconRootsAddress
	case 6:
		return params.WithdrawalQueueAddress
	case 7:
		return params.ConsolidationQueueAddress
	case 8:
		return params.SystemAddress
	case 9:
		return params.HistoryStorageAddress
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

func randomAuthEntry(f *filler.Filler) types.SetCodeAuthorization {
	return types.SetCodeAuthorization{
		ChainID: *uint256.NewInt(f.Uint64()),
		Address: randomAddress(),
		Nonce:   f.Uint64(),
	}
}

func RandomAuthList(f *filler.Filler, sk *ecdsa.PrivateKey) ([]types.SetCodeAuthorization, error) {
	var authList []types.SetCodeAuthorization
	entries := f.MemInt()
	for i := 0; i < int(entries.Uint64()); i++ {
		signed, err := types.SignSetCode(sk, randomAuthEntry(f))
		if err != nil {
			return nil, err
		}
		authList = append(authList, signed)
	}
	return authList, nil
}
