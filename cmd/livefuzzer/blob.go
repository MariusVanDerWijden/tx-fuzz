package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

func SpamBlobTransactions(N uint64, fromCorpus bool, accessList bool, seed int64) {
	backend, _, err := getRealBackend()
	if err != nil {
		log.Warn("Could not get backend", "error", err)
		return
	}
	// Set up the randomness
	random := make([]byte, 10000)
	seed, mut, N := setup(backend, seed, N)

	fmt.Printf("Spamming %v blob transactions per account on %v accounts with seed: 0x%x\n", N, len(keys), seed)
	// Now let everyone spam blob transactions
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
			SendBlobTransactions(backend, sk, f, addr, N, accessList)
		}(keys[i], addrs[i], f)
	}
	wg.Wait()
}

func SendBlobTransactions(client *rpc.Client, key *ecdsa.PrivateKey, f *filler.Filler, addr string, N uint64, al bool) {
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
		tx, err := txfuzz.RandomBlobTx(client, f, sender, nonce, nil, nil, al)
		if err != nil {
			log.Warn("Could not create valid tx: %v", nonce)
			continue
		}
		signedTx, err := types.SignTx(tx.Transaction, types.NewCancunSigner(chainID), key)
		if err != nil {
			panic(err)
		}
		tx.Transaction = signedTx
		rlpData, err := tx.MarshalBinary()
		if err != nil {
			panic(err)
		}
		if err := client.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData)); err != nil {
			if strings.Contains(err.Error(), "account limit exceeded") {
				// Back off for a bit if we send a lot of transactions at once
				time.Sleep(1 * time.Minute)
				continue
			} else {
				panic(err)
			}
		}
		lastTx = signedTx
		time.Sleep(10 * time.Millisecond)
	}

	if lastTx != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Second)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, lastTx); err != nil {
			fmt.Printf("Wait mined failed for blob transactions: %v\n", err.Error())
		}
	}
}
