package main

import "github.com/urfave/cli/v2"

var (
	seedFlag = &cli.Int64Flag{
		Name:  "seed",
		Usage: "Seed for the RNG, (Default = RandomSeed)",
		Value: 0,
	}

	skFlag = &cli.StringFlag{
		Name:  "sk",
		Usage: "Secret key",
		Value: "0xcdfbe6f7602f67a97602e3e9fc24cde1cdffa88acd47745c0b84c5ff55891e1b",
	}

	corpusFlag = &cli.StringFlag{
		Name:  "corpus",
		Usage: "Use additional Corpus",
	}

	noALFlag = &cli.BoolFlag{
		Name:  "no-al",
		Usage: "Disable accesslist creation",
		Value: false,
	}

	countFlag = &cli.IntFlag{
		Name:  "accounts",
		Usage: "Count of accounts to send transactions from",
		Value: 100,
	}

	rpcFlag = &cli.StringFlag{
		Name:  "rpc",
		Usage: "RPC provider",
		Value: "http://127.0.0.1:8545",
	}

	txCountFlag = &cli.IntFlag{
		Name:  "txcount",
		Usage: "Number of transactions send per account per block, 0 = best estimate",
		Value: 0,
	}

	gasLimitFlag = &cli.IntFlag{
		Name:  "gaslimit",
		Usage: "Gas limit used for transactions",
		Value: 100_000,
	}

	spamFlags = []cli.Flag{
		skFlag,
		seedFlag,
		noALFlag,
		corpusFlag,
		rpcFlag,
		txCountFlag,
		countFlag,
		gasLimitFlag,
	}
)
