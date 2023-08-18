package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/inhies/go-bytesize"
)

var (
	address  = "http://127.0.0.1:8545"
	maxSize  = 10 * 1024 * 1024 // 1GB
	gaslimit = uint64(30_000_000)
)

func main() {
	cl, sk := getRealBackend()
	backend := ethclient.NewClient(cl)
	// Set gas limit
	block, err := backend.BlockByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	gaslimit = block.GasLimit()
	// Create transact opts
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(sk, chainid)
	if err != nil {
		panic(err)
	}
	contract := deploy(backend, opts)
	blowUpContract(backend, contract, opts)
	selfdestruct(backend, contract, opts)
}

func blowUpContract(backend *ethclient.Client, contract *Selfdestructer, opts *bind.TransactOpts) {
	size := uint64(0)
	opts.GasLimit = gaslimit / 2
	for i := 0; size < uint64(maxSize); i++ {
		tx, err := contract.Store(opts)
		if err != nil {
			panic(err)
		}
		receipts, err := bind.WaitMined(context.Background(), backend, tx)
		if err != nil {
			panic(err)
		}
		if receipts.Status != types.ReceiptStatusSuccessful {
			panic("status not successful")
		}
		// update size
		newSize, err := contract.Size(&bind.CallOpts{})
		if err != nil {
			panic(err)
		}
		size = newSize.Uint64()
		fmt.Printf("Contract blown up to %v bytes in tx %v\n", bytesize.New(float64(size)), i)
	}
}

func selfdestruct(backend *ethclient.Client, contract *Selfdestructer, opts *bind.TransactOpts) {
	tx, err := contract.Destruct(opts)
	if err != nil {
		panic(err)
	}
	receipts, err := bind.WaitMined(context.Background(), backend, tx)
	if err != nil {
		panic(err)
	}
	if receipts.Status != types.ReceiptStatusSuccessful {
		panic("status not successful")
	}
}

func deploy(backend *ethclient.Client, opts *bind.TransactOpts) *Selfdestructer {
	addr, tx, contract, err := DeploySelfdestructer(opts, backend)
	if err != nil {
		panic(err)
	}
	receipts, err := bind.WaitMined(context.Background(), backend, tx)
	if err != nil {
		panic(err)
	}
	if receipts.Status != types.ReceiptStatusSuccessful {
		panic("status not successful")
	}
	fmt.Printf("Contract deployed at %v\n", addr)
	return contract
}

func getRealBackend() (*rpc.Client, *ecdsa.PrivateKey) {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})

	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
	if crypto.PubkeyToAddress(sk.PublicKey).Hex() != txfuzz.ADDR {
		panic(fmt.Sprintf("wrong address want %s got %s", crypto.PubkeyToAddress(sk.PublicKey).Hex(), txfuzz.ADDR))
	}
	cl, err := rpc.Dial(address)
	if err != nil {
		panic(err)
	}
	return cl, sk
}
