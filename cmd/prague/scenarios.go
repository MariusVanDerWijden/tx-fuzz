package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/MariusVanDerWijden/tx-fuzz/helper"
	"github.com/MariusVanDerWijden/tx-fuzz/spammer"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
)

var nonceOffset = 0

func test7702NormalTxs() {
	fmt.Println("test 7702 normal txs scenario")
	var (
		NumKeys = 512
		numTxs  = 16
		value   = new(big.Int).Lsh(big.NewInt(1), 63)
		addr    = common.HexToAddress(txfuzz.ADDR)
	)
	// set nonce offset
	nonceOffset = numTxs
	keys := spammer.CreateAddressesRaw(NumKeys)
	sendNormalTxs := func(keys []*ecdsa.PrivateKey) {
		cl, _ := helper.GetRealBackend()
		backend := ethclient.NewClient(cl)
		for _, key := range keys {
			sender := crypto.PubkeyToAddress(key.PublicKey)
			nonce, err := backend.PendingNonceAt(context.Background(), sender)
			if err != nil {
				panic(err)
			}
			for i := range numTxs {
				ExecWithSK(backend, key, addr, nonce+uint64(i), nil, false)
			}
		}
	}
	test7702Scenario(keys, value, sendNormalTxs)
}

func test7702BlobTxs() {
	fmt.Println("test 7702 blob txs scenario")
	var (
		NumKeys = 128
		value   = new(big.Int).Lsh(big.NewInt(1), 63)
		addr    = common.HexToAddress(txfuzz.ADDR)
	)
	// set nonce offset
	nonceOffset = 0
	keys := spammer.CreateAddressesRaw(NumKeys)
	sendBlobTxs := func(keys []*ecdsa.PrivateKey) {
		cl, _ := helper.GetRealBackend()
		backend := ethclient.NewClient(cl)
		for _, key := range keys {
			sender := crypto.PubkeyToAddress(key.PublicKey)
			nonce, err := backend.PendingNonceAt(context.Background(), sender)
			if err != nil {
				panic(err)
			}
			ExecWithSK(backend, key, addr, nonce, nil, true)
		}
	}
	test7702Scenario(keys, value, sendBlobTxs)
}

// Test7702Scenario sends a bunch of normal transactions
// and cancels them with a 7702 transaction inducing a bunch
// of churn in the transaction pool.
func test7702Scenario(keys []*ecdsa.PrivateKey, value *big.Int, sendTxs func(keys []*ecdsa.PrivateKey)) {
	// Create configuration
	backend, sk := helper.GetRealBackend()
	config := spammer.NewPartialConfig(backend, sk, keys)

	fmt.Println("Deploying contracts")
	// Deploy the Callee contract
	calleeAddr, err := deploy7702Callee(crypto.PubkeyToAddress(sk.PublicKey).Hex())
	if err != nil {
		panic(err)
	}

	// Deploy the Caller contract
	callerAddr, err := deploy7702Caller()
	if err != nil {
		panic(err)
	}

	// Create the authorizations
	fmt.Println("Creating authorizations")

	// Airdrop the addresses
	fmt.Println("Airdropping")
	if err := spammer.Airdrop(config, value); err != nil {
		panic(err)
	}

	// Send transactions from the accounts
	fmt.Println("Sending transactions")
	sendTxs(keys)
	// Send an auth that invalidates the transactions
	var callsPerTx = min(len(keys), 128)
	var lastTx *types.Transaction
	fmt.Println("Invalidating transactions")
	start := time.Now()
	nonce := helper.Nonce(crypto.PubkeyToAddress(sk.PublicKey))
	for i := range len(keys) / callsPerTx {
		lastTx = sendAuths(keys[i*callsPerTx:i*callsPerTx+callsPerTx], callerAddr, calleeAddr, nonce+uint64(i))
	}
	helper.Wait(lastTx)
	fmt.Printf("Invalidated transactions in %v\n", time.Since(start))
	verify(backend, keys)
}

func sendAuths(keys []*ecdsa.PrivateKey, callerAddr, calleeAddr common.Address, senderNonce uint64) *types.Transaction {
	nonce := uint64(nonceOffset) // hacks
	chainID := uint256.MustFromBig(helper.ChainID())
	auths := authAddress(keys, calleeAddr, chainID, nonce)
	calldata := []byte{}
	for _, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		calldata = append(calldata, addr.Bytes()...)
	}
	return helper.ExecAuthWithNonce(callerAddr, senderNonce, calldata, auths)
}

// verify that the test was executed correctly
func verify(client *rpc.Client, keys []*ecdsa.PrivateKey) {
	backend := ethclient.NewClient(client)
	for _, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		code, err := backend.CodeAt(context.Background(), addr, nil)
		if err != nil {
			panic(err)
		}
		if len(code) == 0 || code[0] != 0xef || code[1] != 0x01 {
			panic(fmt.Sprintf("account %v has no active delegation, has code %v", addr, code))
		}

		balance, err := backend.BalanceAt(context.Background(), addr, nil)
		if err != nil {
			panic(err)
		}
		if balance.Cmp(new(big.Int)) != 0 {
			panic(fmt.Sprintf("account %v was not sweeped, has balance %v", addr, balance))
		}
	}
}

// creates authorizations from keys to addr all with the same nonce.
func authAddress(keys []*ecdsa.PrivateKey, addr common.Address, chainID *uint256.Int, nonce uint64) []types.SetCodeAuthorization {
	auths := make([]types.SetCodeAuthorization, 0, len(keys))
	for _, sk := range keys {
		auth := types.SetCodeAuthorization{
			Address: addr,
			ChainID: *chainID,
			Nonce:   nonce,
		}
		signed, err := types.SignSetCode(sk, auth)
		if err != nil {
			panic(err)
		}
		auths = append(auths, signed)
	}
	return auths
}

func ExecWithSK(backend *ethclient.Client, sk *ecdsa.PrivateKey, addr common.Address, nonce uint64, data []byte, blobs bool) *types.Transaction {
	cl, _ := helper.GetRealBackend()
	sender := crypto.PubkeyToAddress(sk.PublicKey)

	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Nonce: %v\n", nonce)
	gp, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	tip, err := backend.SuggestGasTipCap(context.Background())
	if err != nil {
		panic(err)
	}

	msg := ethereum.CallMsg{
		From:          sender,
		To:            &addr,
		Gas:           uint64(30_000_000),
		GasTipCap:     tip,
		GasFeeCap:     gp,
		Value:         big.NewInt(0),
		Data:          data,
		AccessList:    make(types.AccessList, 0),
		BlobGasFeeCap: big.NewInt(1_000_000),
	}
	if gas, err := backend.EstimateGas(context.Background(), msg); err != nil {
		msg.Gas = uint64(5_000_000)
		fmt.Printf("Error estimating gas: %v, defaulting to %v gas\n", err, msg.Gas)
	} else {
		msg.Gas = gas
	}

	var signedTx *types.Transaction
	if blobs {
		blob, err := helper.RandomBlobData()
		if err != nil {
			panic(err)
		}
		tx := txfuzz.New4844Tx(nonce, msg.To, msg.Gas, chainid, msg.GasTipCap, msg.GasPrice, msg.Value, msg.Data, msg.BlobGasFeeCap, blob, msg.AccessList)
		signedTx, _ = types.SignTx(tx, types.NewCancunSigner(chainid), sk)
	} else {
		tx := types.NewTx(&types.DynamicFeeTx{ChainID: chainid, Nonce: nonce, GasTipCap: msg.GasTipCap, GasFeeCap: msg.GasFeeCap, Gas: msg.Gas, To: msg.To, Data: msg.Data, Value: msg.Value, AccessList: msg.AccessList})
		signedTx, _ = types.SignTx(tx, types.NewCancunSigner(chainid), sk)
	}

	rlpData, err := signedTx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	if err := cl.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData)); err != nil {
		fmt.Println(err)
	}
	return signedTx
}
