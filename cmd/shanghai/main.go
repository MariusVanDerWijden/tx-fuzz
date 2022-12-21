package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	address = "http://127.0.0.1:8545"
)

func main() {
	// Store coinbase
	exec([]byte{0x41, 0x41, 0x55})
	// Call coinbase
	// 5x PUSH0, COINBASE, GAS, CALL
	exec([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xf1})
	// 5x PUSH0, COINBASE, GAS, CALLCODE
	exec([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xf2})
	// 5x PUSH0, COINBASE, GAS, DELEGATECALL
	exec([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xf4})
	// 5x PUSH0, COINBASE, GAS, STATICCALL
	exec([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x41, 0x5A, 0xfA})
	// COINBASE, SELFDESTRUCT
	exec([]byte{0x41, 0xff})
	// COINBASE, EXTCODESIZE
	exec([]byte{0x41, 0x3b})
	// 3x PUSH0, COINBASE, EXTCODECOPY
	exec([]byte{0x5f, 0x5f, 0x5f, 0x41, 0x3C})
	// COINBASE, EXTCODEHASH
	exec([]byte{0x41, 0x3F})
	// COINBASE, BALANCE
	exec([]byte{0x41, 0x31})
	// loop push0
	// JUMPDEST, PUSH0, JUMP
	exec([]byte{0x58, 0x5f, 0x56})
}

func exec(data []byte) {
	cl, sk := getRealBackend()
	backend := ethclient.NewClient(cl)
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		panic(err)
	}
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Nonce: %v\n", nonce)
	gp, _ := backend.SuggestGasPrice(context.Background())
	tx := types.NewContractCreation(nonce, common.Big1, 500000, gp, data)
	signedTx, _ := types.SignTx(tx, types.NewLondonSigner(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
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
