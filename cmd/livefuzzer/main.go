package main

import (
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
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
	Action: runBasicSpam,
	Flags: []cli.Flag{
		skFlag,
		seedFlag,
		noALFlag,
		corpusFlag,
		rpcFlag,
		txCountFlag,
	},
}

var blobSpamCommand = &cli.Command{
	Name:   "blobs",
	Usage:  "Send blob spam transactions",
	Action: runBlobSpam,
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
		blobSpamCommand,
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

func spam(c *cli.Context, basic bool) error {
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
		if basic {
			SpamBasicTransactions(uint64(txPerAccount), false, !noAL, seed)
		} else {
			SpamBlobTransactions(uint64(txPerAccount), false, !noAL, seed)
		}
		time.Sleep(12 * time.Second)
	}
}

func runBasicSpam(c *cli.Context) error {
	return spam(c, true)
}

func runBlobSpam(c *cli.Context) error {
	return spam(c, false)
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
