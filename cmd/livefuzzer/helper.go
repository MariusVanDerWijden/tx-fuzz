package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	txfuzz "github.com/mieubrisse/tx-fuzz"
)

func getRealBackend(address string) (*rpc.Client, *ecdsa.PrivateKey) {
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

func sendTx(sk *ecdsa.PrivateKey, backend *ethclient.Client, to common.Address, value *big.Int) {
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Nonce: %v\n", nonce)
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	gp, _ := backend.SuggestGasPrice(context.Background())
	tx := types.NewTransaction(nonce, to, value, 500000, gp, nil)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
}

func unstuck(sk *ecdsa.PrivateKey, backend *ethclient.Client, sender, to common.Address, value, gasPrice *big.Int) {
	blocknumber, err := backend.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(int64(blocknumber)))
	if err != nil {
		panic(err)
	}
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Acc: %v Nonce: %v\n", sender, nonce)
	if gasPrice == nil {
		gasPrice, _ = backend.SuggestGasPrice(context.Background())
	}
	tx := types.NewTransaction(nonce, to, value, 21000, gasPrice, nil)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
}
