package main

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/urfave/cli/v2"
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
	Flags:  spamFlags,
}

var blobSpamCommand = &cli.Command{
	Name:   "blobs",
	Usage:  "Send blob spam transactions",
	Action: runBlobSpam,
	Flags:  spamFlags,
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

func unstuckTransactions(config *Config) {
	client := ethclient.NewClient(config.backend)
	faucetAddr := crypto.PubkeyToAddress(config.faucet.PublicKey)
	var wg sync.WaitGroup
	wg.Add(len(config.keys))
	for _, key := range config.keys {
		go func(key *ecdsa.PrivateKey) {
			unstuck(config.faucet, client, faucetAddr, common.Big0, nil)
			wg.Done()
		}(key)
	}
	wg.Wait()
}

func runAirdrop(c *cli.Context) error {
	config, err := NewConfigFromContext(c)
	if err != nil {
		return err
	}
	txPerAccount := config.n
	airdropValue := new(big.Int).Mul(big.NewInt(int64(txPerAccount*100000)), big.NewInt(params.GWei))
	airdrop(config, airdropValue)
	return nil
}

func spam(config *Config, spamFn Spam, airdropValue *big.Int) error {
	for {
		if err := airdrop(config, airdropValue); err != nil {
			return err
		}
		SpamTransactions(config, spamFn)
		time.Sleep(12 * time.Second)
	}
}

func runBasicSpam(c *cli.Context) error {
	config, err := NewConfigFromContext(c)
	if err != nil {
		return err
	}
	airdropValue := new(big.Int).Mul(big.NewInt(int64((1+config.n)*1000000)), big.NewInt(params.GWei))
	return spam(config, SendBasicTransactions, airdropValue)
}

func runBlobSpam(c *cli.Context) error {
	config, err := NewConfigFromContext(c)
	if err != nil {
		return err
	}
	airdropValue := new(big.Int).Mul(big.NewInt(int64((1+config.n)*1000000)), big.NewInt(params.GWei))
	airdropValue = airdropValue.Mul(airdropValue, big.NewInt(100))
	return spam(config, SendBasicTransactions, airdropValue)
}

func runUnstuck(c *cli.Context) error {
	config, err := NewConfigFromContext(c)
	if err != nil {
		return err
	}
	unstuckTransactions(config)
	return nil
}

func runCreate(c *cli.Context) error {
	createAddresses(100)
	return nil
}
