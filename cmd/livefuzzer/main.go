package main

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	txfuzz "github.com/mariusvanderwijden/tx-fuzz"
)

var (
	address      = "http://127.0.0.1:8545"
	txPerAccount = 1000
	airdropValue = new(big.Int).Mul(big.NewInt(int64(txPerAccount*10000)), big.NewInt(params.GWei))
)

func main() {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})
	if len(os.Args) != 2 {
		panic("invalid amount of args, need 2")
	}
	switch os.Args[1] {
	case "airdrop":
		airdrop(airdropValue)
	case "spam":
		SpamTransactions(uint64(txPerAccount))
	default:
		fmt.Println("unrecognized option")
	}
}

func SpamTransactions(N uint64) {
	backend, _ := getRealBackend()
	// Now let everyone spam baikal transactions
	var wg sync.WaitGroup
	wg.Add(len(keys))
	for i, key := range keys {
		go func(key, addr string) {
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			SendBaikalTransactions(backend, sk, addr, N)
			wg.Done()
		}(key, addrs[i])
	}
	wg.Wait()
}

func SendBaikalTransactions(client *rpc.Client, key *ecdsa.PrivateKey, addr string, N uint64) {
	backend := ethclient.NewClient(client)
	rnd := make([]byte, 10000)
	crand.Read(rnd)
	f := filler.NewFiller(rnd)

	sender := common.HexToAddress(addr)
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	nonce, err := backend.NonceAt(context.Background(), sender, nil)
	if err != nil {
		panic(err)
	}
	for i := uint64(0); i < N; i++ {

		tx, err := txfuzz.RandomValidTx(client, f, sender, nonce+i, nil, nil)
		if err != nil {
			fmt.Print(err)
			continue
		}
		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainid), key)
		if err != nil {
			panic(err)
		}
		err = backend.SendTransaction(context.Background(), signedTx)
		if err == nil {
			nonce++
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, signedTx); err != nil {
			fmt.Printf("Wait mined failed: %v\n", err.Error())
		}
	}
}
