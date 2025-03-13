package spammer

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

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
	backends []*rpc.Client // connection to the rpc provider

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

func NewDefaultConfig(rpcAddrs string, N uint64, accessList bool, rng *rand.Rand) (*Config, error) {
	// Setup RPC
	backends := make([]*rpc.Client, 0)

	for _, rpcAddr := range strings.Split(rpcAddrs, ",") {
		backend, err := rpc.Dial(rpcAddr)
		if err == nil {
			backends = append(backends, backend)
		}
	}

	if len(backends) == 0 {
		return nil, errors.New("no valid rpc addresses found")
	}

	// Setup Keys
	var keys []*ecdsa.PrivateKey
	for i := 0; i < len(staticKeys); i++ {
		keys = append(keys, crypto.ToECDSAUnsafe(common.FromHex(staticKeys[i])))
	}

	return &Config{
		backends:   backends,
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

func NewConfigFromContext(c *cli.Context) (*Config, error) {
	// Setup RPC
	var rpcAddrs []string
	if rpcs := c.String(flags.RpcsFlag.Name); rpcs != "" {
		rpcAddrs = strings.Split(rpcs, ",")
	} else {
		rpcAddrs = []string{c.String(flags.RpcFlag.Name)}
	}

	backends := make([]*rpc.Client, 0)
	for _, rpcAddr := range rpcAddrs {
		backend, err := rpc.Dial(rpcAddr)
		if err == nil {
			backends = append(backends, backend)
		}
	}

	if len(backends) == 0 {
		return nil, errors.New("no valid rpc addresses found")
	}

	// Setup faucet
	faucet := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
	var err error
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
		N, err = setupN(backends[0], len(keys), gasLimit)
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
		backends:   backends,
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
