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
	"github.com/MariusVanDerWijden/tx-fuzz/mutator"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

var (
	address    = "http://127.0.0.1:8545"
	corpus     [][]byte
	defaultGas = 100000
)

var airdropCommand = &cli.Command{
	Name:   "airdrop",
	Usage:  "Airdrops to a list of accounts",
	Action: runAirdrop,
	Flags: []cli.Flag{
		skFlag,
		rpcFlag,
	},
}

var spamCommand = &cli.Command{
	Name:   "spam",
	Usage:  "Send spam transactions",
	Action: runSpam,
	Flags: []cli.Flag{
		skFlag,
		seedFlag,
		noALFlag,
		corpusFlag,
		rpcFlag,
		txCountFlag,
	},
}

var unstuckCommand = &cli.Command{
	Name:   "unstuck",
	Usage:  "Tries to unstuck an account",
	Action: runUnstuck,
	Flags: []cli.Flag{
		skFlag,
		rpcFlag,
	},
}
var sendCommand = &cli.Command{
	Name:   "send",
	Usage:  "Sends a single transaction",
	Action: runSend,
	Flags: []cli.Flag{
		skFlag,
		rpcFlag,
	},
}
var createCommand = &cli.Command{
	Name:   "create",
	Usage:  "Create ephemeral accounts",
	Action: runCreate,
	Flags: []cli.Flag{
		countFlag,
		rpcFlag,
	},
}

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = "tx-fuzz"
	app.Usage = "Fuzzer for sending spam transactions"
	app.Commands = []*cli.Command{
		airdropCommand,
		spamCommand,
		unstuckCommand,
		sendCommand,
		createCommand,
	}
	return app
}

var app = initApp()

func main() {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func SpamTransactions(N uint64, fromCorpus bool, accessList bool, seed int64) {
	backend, _, err := getRealBackend()
	if err != nil {
		log.Warn("Could not get backend", "error", err)
		return
	}
	// Setup seed
	var src *rand.Rand
	if seed == 0 {
		fmt.Println("No seed provided, creating one")
		rnd := make([]byte, 8)
		crand.Read(rnd)
		s := int64(binary.BigEndian.Uint64(rnd))
		seed = s
	}
	src = rand.New(rand.NewSource(seed))
	mut := mutator.NewMutator(src)
	// Set up the randomness
	random := make([]byte, 10000)
	// Setup N
	if N == 0 {
		client := ethclient.NewClient(backend)
		block, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			panic(err)
		}
		txPerBlock := block.GasLimit() / uint64(defaultGas)
		txPerAccount := txPerBlock / uint64(len(keys))
		N = txPerAccount
		if N == 0 {
			N = 1
		}
	}
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
		go func(key, addr string, filler *filler.Filler, index int) {
			defer wg.Done()
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			if index < len(keys)/10 {
				SendBlobTransactions(backend, sk, f, addr, N/10, accessList) // Send blob txs with one tenth of accounts
			} else {
				SendBaikalTransactions(backend, sk, f, addr, N, accessList)
			}
		}(keys[i], addrs[i], f, i)
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
		signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainid), key)
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

func SendBlobTransactions(client *rpc.Client, key *ecdsa.PrivateKey, f *filler.Filler, addr string, N uint64, al bool) {
	backend := ethclient.NewClient(client)

	sender := common.HexToAddress(addr)
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		log.Warn("Could not get chainid, using default")
		chainid = big.NewInt(0x01000666)
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
		signedTx, err := types.SignTx(tx.Transaction, types.NewCancunSigner(chainid), key)
		if err != nil {
			panic(err)
		}
		tx.Transaction = signedTx
		rlpData, err := tx.MarshalBinary()
		if err != nil {
			panic(err)
		}
		if err := client.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData)); err != nil {
			panic(err)
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

func runAirdrop(c *cli.Context) error {
	setupDefaults(c)
	txPerAccount := 10000
	airdropValue := new(big.Int).Mul(big.NewInt(int64(txPerAccount*100000)), big.NewInt(params.GWei))
	airdrop(airdropValue)
	return nil
}

func runSpam(c *cli.Context) error {
	setupDefaults(c)
	noAL := c.Bool(noALFlag.Name)
	seed := c.Int64(seedFlag.Name)
	txPerAccount := c.Int(txCountFlag.Name)
	// Setup corpus if needed
	if corpusFile := c.String(corpusFlag.Name); corpusFile != "" {
		cp, err := readCorpusElements(corpusFile)
		if err != nil {
			panic(err)
		}
		corpus = cp
	}
	// Limit amount of accounts
	keys = keys[:10]
	addrs = addrs[:10]

	for {
		airdropValue := new(big.Int).Mul(big.NewInt(int64((1+txPerAccount)*1000000)), big.NewInt(params.GWei))
		if err := airdrop(airdropValue); err != nil {
			return err
		}
		SpamTransactions(uint64(txPerAccount), false, !noAL, seed)
		time.Sleep(12 * time.Second)
	}
}

func runUnstuck(c *cli.Context) error {
	setupDefaults(c)
	unstuckTransactions()
	return nil
}

func runSend(c *cli.Context) error {
	setupDefaults(c)
	send()
	return nil
}

func runCreate(c *cli.Context) error {
	setupDefaults(c)
	createAddresses(100)
	return nil
}

func setupDefaults(c *cli.Context) {
	if sk := c.String(skFlag.Name); sk != "" {
		txfuzz.SK = sk
		sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
		txfuzz.ADDR = crypto.PubkeyToAddress(sk.PublicKey).Hex()
	}
	address = c.String(rpcFlag.Name)
}
