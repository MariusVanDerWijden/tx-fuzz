package main

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"strings"
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
	txfuzz "github.com/mieubrisse/tx-fuzz"
)

const (
	numSpammingThreads = 10
)

var (
	txPerAccount = 1000
	airdropValue = new(big.Int).Mul(big.NewInt(int64(txPerAccount*100000)), big.NewInt(params.GWei))
	corpus       [][]byte
)

func main() {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})
	if len(os.Args) < 3 {
		panic(fmt.Sprintf("Usage: %v [node_ip:rpc_port] [command]", os.Args[0]))
	}

	rpcUrl := os.Args[1]

	switch os.Args[2] {
	case "airdrop":
		panic("TODO Can't airdrop on generic networks until the faucet account is parameterized")
		airdrop(rpcUrl, airdropValue)
	case "spam":
		// The private keys of the addresses that will send transactions
		commaSeparatedPrivateKeys := os.Args[3]
		// The addresses that the private keys correspond to
		commaSeparatedAddresses := os.Args[4]
		SpamTransactions(rpcUrl, commaSeparatedPrivateKeys, commaSeparatedAddresses, false)
	case "corpus":
		// The private keys of the addresses that will send transactions
		commaSeparatedPrivateKeys := os.Args[3]
		// The addresses that the private keys correspond to
		commaSeparatedAddresses := os.Args[4]
		cp, err := readCorpusElements(os.Args[5])
		if err != nil {
			panic(err)
		}
		corpus = cp
		SpamTransactions(rpcUrl, commaSeparatedPrivateKeys, commaSeparatedAddresses, true)
	case "unstuck":
		unstuckTransactions(rpcUrl)
	case "send":
		send(rpcUrl)
	default:
		fmt.Println("unrecognized option")
	}
}

func SpamTransactions(rpcUrl string, commaSeparatedPrivateKeys string, commaSeparatedAddresses string, fromCorpus bool) {
	backend, _ := getRealBackend(rpcUrl)

	privateKeyStrs := strings.Split(commaSeparatedPrivateKeys, ",")
	addressStrs := strings.Split(commaSeparatedPrivateKeys, ",")

	privateKeys := []*ecdsa.PrivateKey{}
	for _, keyStr := range privateKeyStrs {
		key := crypto.ToECDSAUnsafe(common.FromHex(keyStr))
		privateKeys = append(privateKeys, key)
	}

	for i := 0; i < numSpammingThreads; i++ {
		go func() {
			var f *filler.Filler
			if fromCorpus {
				elem := corpus[rand.Int31n(int32(len(corpus)))]
				f = filler.NewFiller(elem)
			} else {
				rnd := make([]byte, 10000)
				crand.Read(rnd)
				f = filler.NewFiller(rnd)
			}
			SendBaikalTransactions(backend, privateKeys, f, addressStrs)
		}()
	}
}

// Repeatedly sends transactions from a random source to a random destination
func SendBaikalTransactions(client *rpc.Client, keys []*ecdsa.PrivateKey, f *filler.Filler, addresses []string) {
	backend := ethclient.NewClient(client)

	// Pick a random source to send ETH from
	idx := rand.Intn(len(keys))
	key := keys[idx]
	srcAddr := addresses[idx]

	sender := common.HexToAddress(srcAddr)
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	for {
		nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(-1))
		if err != nil {
			panic(err)
		}
		tx, err := txfuzz.RandomValidTx(client, f, sender, nonce, nil, nil)
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

func unstuckTransactions(rpcUrl string) {
	backend, _ := getRealBackend(rpcUrl)
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
	stats, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	corpus := make([][]byte, 0, len(stats))
	for _, file := range stats {
		b, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", path, file.Name()))
		if err != nil {
			return nil, err
		}
		corpus = append(corpus, b)
	}
	return corpus, nil
}

func send(rpcUrl string) {
	backend, _ := getRealBackend(rpcUrl)
	client := ethclient.NewClient(backend)
	to := common.HexToAddress(txfuzz.ADDR)
	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK2))
	value := new(big.Int).Mul(big.NewInt(100000), big.NewInt(params.Ether))
	sendTx(sk, client, to, value)
}
