package main

import (
	"context"
	"fmt"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/MariusVanDerWijden/tx-fuzz/helper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	cl, sk := helper.GetRealBackend()
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
	tx := types.NewContractCreation(nonce, common.Big1, 500000, gp, []byte{0x44, 0x44, 0x55})
	signedTx, _ := types.SignTx(tx, types.NewLondonSigner(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
}
