package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	maxDataPerTx = 1 << 17 // 128Kb
)

var (
	address = "http://127.0.0.1:8545"
)

func main() {
	// PUSH0, DATAHASH, PUSH0, DATAHASH, SSTORE
	exec([]byte{0x5f, 0x49, 0x5f, 0x49, 0x55})
}

func exec(data []byte) {
	cl, sk := getRealBackend()
	backend := ethclient.NewClient(cl)
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		panic(err)
	}
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Nonce: %v\n", nonce)
	gp, _ := backend.SuggestGasPrice(context.Background())
	tip, _ := backend.SuggestGasTipCap(context.Background())
	blob, _ := randomBlobData()
	nonce = nonce - 2
	tx := txfuzz.New4844Tx(nonce, nil, 500000, chainid, tip.Mul(tip, common.Big1), gp.Mul(gp, common.Big1), common.Big1, data, blob, make(types.AccessList, 0))
	signedTx, _ := types.SignTx(tx, types.NewDankSigner(chainid), sk)
	if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
		panic(err)
	}
}

func getRealBackend() (*rpc.Client, *ecdsa.PrivateKey) {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})

	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
	if crypto.PubkeyToAddress(sk.PublicKey).Hex() != txfuzz.ADDR {
		panic(fmt.Sprintf("wrong address want %s got %s", crypto.PubkeyToAddress(sk.PublicKey).Hex(), txfuzz.ADDR))
	}
	cl, err := rpc.Dial(address)
	if err != nil {
		panic(err)
	}
	return cl, sk
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
