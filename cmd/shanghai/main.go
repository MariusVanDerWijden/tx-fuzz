package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"time"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	address         = "http://127.0.0.1:8545"
	maxCodeSize     = 24576
	maxInitCodeSize = 2 * maxCodeSize
)

func main() {
	for {
		runTests()
		time.Sleep(time.Minute)
	}
}

func runTests() {
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
	fmt.Println("Limit&MeterInitcode")
	// limit & meter initcode
	sizes := []int{
		maxInitCodeSize - 2,
		maxInitCodeSize - 1,
		maxInitCodeSize,
		maxInitCodeSize + 1,
		maxInitCodeSize + 2,
		maxInitCodeSize * 2,
	}
	// PUSH4 size, PUSH0, PUSH0, CREATE
	for _, size := range sizes {
		initcode := pushSize(size)
		exec(append(initcode, []byte{0x57, 0x57, 0xF0}...))
	}
	// PUSH4 size, PUSH0, PUSH0, CREATE2
	for _, size := range sizes {
		initcode := pushSize(size)
		exec(append(initcode, []byte{0x57, 0x57, 0xF5}...))
	}
	// size x JUMPDEST STOP
	for _, size := range sizes {
		initcode := repeatOpcode(size, 0x58)
		exec(append(initcode, 0x00))
	}
	// size x STOP STOP
	for _, size := range sizes {
		exec(repeatOpcode(size, 0x00))
	}
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
	tx := types.NewContractCreation(nonce, common.Big1, 5000000, gp.Mul(gp, common.Big2), data)
	signedTx, _ := types.SignTx(tx, types.NewLondonSigner(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
}

// PUSH4 size
func pushSize(size int) []byte {
	code := []byte{63}
	sizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeArr, uint32(size))
	code = append(code, sizeArr...)
	return code
}

func repeatOpcode(size int, opcode byte) []byte {
	initcode := []byte{}
	for i := 0; i < size; i++ {
		initcode = append(initcode, opcode)
	}
	return initcode
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
