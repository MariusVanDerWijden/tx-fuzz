package spammer

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/MariusVanDerWijden/tx-fuzz/flags"
	"github.com/MariusVanDerWijden/tx-fuzz/mutator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

type Config struct {
	backend *rpc.Client // connection to the rpc provider

	N          uint64              // number of transactions send per account
	faucet     *ecdsa.PrivateKey   // private key of the faucet account
	keys       []*ecdsa.PrivateKey // private keys of accounts
	corpus     [][]byte            // optional corpus to use elements from
	accessList bool                // whether to create accesslist transactions
	gasLimit   uint64              // gas limit per transaction
	SlotTime   uint64              // slot time in seconds

	seed int64            // seed used for generating randomness
	mut  *mutator.Mutator // Mutator based on the seed
}

func NewDefaultConfig(rpcAddr string, N uint64, accessList bool, rng *rand.Rand) (*Config, error) {
	// Setup RPC
	backend, err := rpc.Dial(rpcAddr)
	if err != nil {
		return nil, err
	}

	// Setup Keys
	var keys []*ecdsa.PrivateKey
	for i := 0; i < len(staticKeys); i++ {
		keys = append(keys, crypto.ToECDSAUnsafe(common.FromHex(staticKeys[i])))
	}

	return &Config{
		backend:    backend,
		N:          N,
		faucet:     crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK)),
		keys:       keys,
		corpus:     [][]byte{},
		accessList: accessList,
		gasLimit:   30_000_000,
		seed:       0,
		mut:        mutator.NewMutator(rng),
	}, nil
}

func NewPartialConfig(backend *rpc.Client, faucet *ecdsa.PrivateKey, keys []*ecdsa.PrivateKey) *Config {
	return &Config{
		backend: backend,
		faucet:  faucet,
		keys:    keys,
	}
}

func NewConfigFromContext(c *cli.Context) (*Config, error) {
	// Setup RPC
	rpcAddr := c.String(flags.RpcFlag.Name)
	backend, err := rpc.Dial(rpcAddr)
	if err != nil {
		return nil, err
	}

	// Setup faucet
	faucet := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
	if sk := c.String(flags.SkFlag.Name); sk != "" {
		faucet, err = crypto.ToECDSA(common.FromHex(sk))
		if err != nil {
			return nil, err
		}
	}

	// Setup Keys
	var keys []*ecdsa.PrivateKey
	nKeys := c.Int(flags.CountFlag.Name)
	if nKeys == 0 || nKeys > len(staticKeys) {
		fmt.Printf("Sanitizing count flag from %v to %v\n", nKeys, len(staticKeys))
		nKeys = len(staticKeys)
	}
	for i := 0; i < nKeys; i++ {
		keys = append(keys, crypto.ToECDSAUnsafe(common.FromHex(staticKeys[i])))
	}

	// Setup gasLimit
	gasLimit := c.Int(flags.GasLimitFlag.Name)

	// Setup N
	N := c.Int(flags.TxCountFlag.Name)
	if N == 0 {
		N, err = setupN(backend, len(keys), gasLimit)
		if err != nil {
			return nil, err
		}
	}

	slotTime := c.Uint64(flags.SlotTimeFlag.Name)

	// Setup seed
	seed := c.Int64(flags.SeedFlag.Name)
	if seed == 0 {
		rnd := make([]byte, 8)
		crand.Read(rnd)
		seed = int64(binary.BigEndian.Uint64(rnd))
		fmt.Printf("No seed provided, creating one: %x\n", seed)
	}

	// Setup Mutator
	mut := mutator.NewMutator(rand.New(rand.NewSource(seed)))

	// Setup corpus
	var corpus [][]byte
	if corpusFile := c.String(flags.CorpusFlag.Name); corpusFile != "" {
		corpus, err = readCorpusElements(corpusFile)
		if err != nil {
			return nil, err
		}
	}

	return &Config{
		backend:    backend,
		N:          uint64(N),
		faucet:     faucet,
		accessList: !c.Bool(flags.NoALFlag.Name),
		gasLimit:   uint64(gasLimit),
		seed:       seed,
		keys:       keys,
		corpus:     corpus,
		mut:        mut,
		SlotTime:   slotTime,
	}, nil
}

func setupN(backend *rpc.Client, keys int, gasLimit int) (int, error) {
	client := ethclient.NewClient(backend)
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	txPerBlock := int(header.GasLimit / uint64(gasLimit))
	txPerAccount := txPerBlock / keys
	if txPerAccount == 0 {
		return 1, nil
	}
	return txPerAccount, nil
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
