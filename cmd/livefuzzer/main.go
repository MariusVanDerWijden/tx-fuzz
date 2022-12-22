package main

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	address      = "http://127.0.0.1:8545"
	txPerAccount = 1000
	airdropValue = new(big.Int).Mul(big.NewInt(int64(txPerAccount*100000)), big.NewInt(params.GWei))
	corpus       [][]byte
)

func main() {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})
	if len(os.Args) < 2 {
		panic("invalid amount of args, need 2")
	}

	var (
		accesslist = true
		seed       *int64
	)

	var offset int
	if len(os.Args) >= 3 {
		a := common.LeftPadBytes(common.FromHex(os.Args[2]), 8)
		s := int64(binary.BigEndian.Uint64(a))
		seed = &s
		offset += 1
	}

	if len(os.Args) > 3+offset && os.Args[3+offset] == "no-al" {
		accesslist = false
		offset += 1
	}

	if len(os.Args) > 3+offset {
		txfuzz.SK = os.Args[3+offset]
		sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
		txfuzz.ADDR = crypto.PubkeyToAddress(sk.PublicKey).Hex()
	}

	if len(os.Args) > 4+offset {
		address = os.Args[4+offset]
	}

	switch os.Args[1] {
	case "airdrop":
		airdrop(airdropValue)
	case "spam":
		for {
			airdrop(airdropValue)
			SpamTransactions(uint64(txPerAccount), false, accesslist, seed)
			time.Sleep(10 * time.Second)
		}
	case "corpus":
		cp, err := readCorpusElements(os.Args[2])
		if err != nil {
			panic(err)
		}
		corpus = cp
		SpamTransactions(uint64(txPerAccount), true, accesslist, seed)
	case "unstuck":
		unstuckTransactions()
	case "send":
		send()
	case "create":
		createAddresses(100)
	default:
		fmt.Println("unrecognized option")
	}
}

func SpamTransactions(N uint64, fromCorpus bool, accessList bool, seed *int64) {
	backend, _, err := getRealBackend()
	if err != nil {
		log.Warn("Could not get backend", "error", err)
		return
	}
	var src rand.Rand
	if seed == nil {
		fmt.Println("No seed provided, creating one")
		rnd := make([]byte, 8)
		crand.Read(rnd)
		s := int64(binary.BigEndian.Uint64(rnd))
		seed = &s
	}
	src = *rand.New(rand.NewSource(*seed))
	fmt.Printf("Spamming transactions with seed: 0x%x\n", *seed)
	// Now let everyone spam baikal transactions
	var wg sync.WaitGroup
	wg.Add(len(keys))
	for i, key := range keys {
		// Set up the randomness
		random := make([]byte, 10000)
		src.Read(random)
		var f *filler.Filler
		if fromCorpus {
			elem := corpus[rand.Int31n(int32(len(corpus)))]
			f = filler.NewFiller(elem)
		} else {
			f = filler.NewFiller(random)
		}
		// Start a fuzzing thread
		go func(key, addr string, filler *filler.Filler) {
			defer wg.Done()
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			SendBaikalTransactions(backend, sk, f, addr, N, accessList)
		}(key, addrs[i], f)
	}
	wg.Wait()
}

func SendBaikalTransactions(client *rpc.Client, key *ecdsa.PrivateKey, f *filler.Filler, addr string, N uint64, al bool) {
	backend := ethclient.NewClient(client)

	sender := common.HexToAddress(addr)
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		log.Warn("Could not get chainid, using default")
		chainid = big.NewInt(0x01000666)
	}

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
		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainid), key)
		if err != nil {
			panic(err)
		}
		backend.SendTransaction(context.Background(), signedTx)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, signedTx); err != nil {
			fmt.Printf("Wait mined failed: %v\n", err.Error())
		}
	}
}

func unstuckTransactions() {
	backend, _, err := getRealBackend()
	if err != nil {
		log.Warn("Could not get backend", "error", err)
		return
	}
	client := ethclient.NewClient(backend)
	// Now let everyone spam baikal transactions
	var wg sync.WaitGroup
	wg.Add(len(keys))
	for i, key := range keys {
		go func(key, addr string) {
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			unstuck(sk, client, common.HexToAddress(addr), common.HexToAddress(addr), common.Big0, nil)
			wg.Done()
		}(key, addrs[i])
	}
	wg.Wait()
}

func readCorpusElements(path string) ([][]byte, error) {
	stats, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	corpus := make([][]byte, 0, len(stats))
	for _, file := range stats {
		b, err := os.ReadFile(fmt.Sprintf("%v/%v", path, file.Name()))
		if err != nil {
			return nil, err
		}
		corpus = append(corpus, b)
	}
	return corpus, nil
}

func send() {
	backend, _, _ := getRealBackend()
	client := ethclient.NewClient(backend)
	to := common.HexToAddress(txfuzz.ADDR)
	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK2))
	value := new(big.Int).Mul(big.NewInt(100000), big.NewInt(params.Ether))
	sendTx(sk, client, to, value)
}
