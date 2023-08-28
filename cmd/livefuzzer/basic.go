package main

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/MariusVanDerWijden/tx-fuzz/mutator"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

func setup(backend *rpc.Client, seed int64, N uint64) (int64, *mutator.Mutator, uint64) {
	// Setup seed
	if seed == 0 {
		fmt.Println("No seed provided, creating one")
		rnd := make([]byte, 8)
		crand.Read(rnd)
		s := int64(binary.BigEndian.Uint64(rnd))
		seed = s
	}

	// Setup N
	client := ethclient.NewClient(backend)
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	if N == 0 {
		txPerBlock := header.GasLimit / uint64(defaultGas)
		txPerAccount := txPerBlock / uint64(len(keys))
		N = txPerAccount
		if N == 0 {
			N = 1
		}
	}

	mut := mutator.NewMutator(rand.New(rand.NewSource(seed)))
	return seed, mut, N
}

func SpamBasicTransactions(N uint64, fromCorpus bool, accessList bool, seed int64) {
	backend, _, err := getRealBackend()
	if err != nil {
		log.Warn("Could not get backend", "error", err)
		return
	}

	// Set up the randomness
	random := make([]byte, 10000)
	seed, mut, N := setup(backend, seed, N)

	fmt.Printf("Spamming %v transactions per account on %v accounts with seed: 0x%x\n", N, len(keys), seed)
	// Now let everyone spam baikal transactions
	var wg sync.WaitGroup
	wg.Add(len(keys))
	for i := range keys {
		var f *filler.Filler
		if fromCorpus {
			elem := corpus[rand.Int31n(int32(len(corpus)))]
			mut.MutateBytes(&elem)
			f = filler.NewFiller(elem)
		} else {
			// Use lower entropy randomness for filler
			mut.MutateBytes(&random)
			f = filler.NewFiller(random)
		}
		// Start a fuzzing thread
		go func(key, addr string, filler *filler.Filler) {
			defer wg.Done()
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			SendBasicTransactions(backend, sk, f, addr, N, accessList)
		}(keys[i], addrs[i], f)
	}
	wg.Wait()
}

func SendBasicTransactions(client *rpc.Client, key *ecdsa.PrivateKey, f *filler.Filler, addr string, N uint64, al bool) {
	backend := ethclient.NewClient(client)

	sender := common.HexToAddress(addr)
	chainID, err := backend.ChainID(context.Background())
	if err != nil {
		log.Warn("Could not get chainID, using default")
		chainID = big.NewInt(0x01000666)
	}

	var lastTx *types.Transaction
	for i := uint64(0); i < N; i++ {
		nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(-1))
		if err != nil {
			log.Warn("Could not get nonce: %v", nonce)
			continue
		}
		tx, err := txfuzz.RandomValidTx(client, f, sender, nonce, nil, nil, al)
		if err != nil {
			log.Warn("Could not create valid tx: %v", nonce)
			continue
		}
		signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainID), key)
		if err != nil {
			panic(err)
		}
		if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
			log.Warn("Could not submit transaction: %v", err)
			continue
		}
		lastTx = signedTx
		time.Sleep(10 * time.Millisecond)
	}
	if lastTx != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Second)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, lastTx); err != nil {
			fmt.Printf("Wait mined failed for SendBaikalTransactions: %v\n", err.Error())
		}
	}
}
