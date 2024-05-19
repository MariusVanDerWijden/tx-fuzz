package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	maxDataPerBlob = 1 << 17 // 128Kb
)

func exec(addr common.Address, data []byte, blobs bool) *types.Transaction {
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
	gp, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	tip, err := backend.SuggestGasTipCap(context.Background())
	if err != nil {
		tip = big.NewInt(100000000)
		//panic(err)
	}
	var rlpData []byte
	var _tx *types.Transaction
	gasLimit := uint64(30_000_000)
	if blobs {
		blob, err := randomBlobData()
		if err != nil {
			panic(err)
		}
		//nonce = nonce - 2
		tx := txfuzz.New4844Tx(nonce, &addr, gasLimit, chainid, tip.Mul(tip, common.Big1), gp.Mul(gp, common.Big1), common.Big0, data, big.NewInt(1_000_000), blob, make(types.AccessList, 0))
		signedTx, _ := types.SignTx(tx, types.NewCancunSigner(chainid), sk)
		rlpData, err = signedTx.MarshalBinary()
		if err != nil {
			panic(err)
		}
		_tx = signedTx
	} else {
		tx := types.NewTx(&types.DynamicFeeTx{ChainID: chainid, Nonce: nonce, GasTipCap: tip, GasFeeCap: gp, Gas: gasLimit, To: &addr, Data: data})
		signedTx, _ := types.SignTx(tx, types.NewCancunSigner(chainid), sk)
		rlpData, err = signedTx.MarshalBinary()
		if err != nil {
			panic(err)
		}
		_tx = signedTx
	}

	if err := cl.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData)); err != nil {
		panic(err)
	}
	return _tx
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

func wait(tx *types.Transaction) {
	client, _ := getRealBackend()
	backend := ethclient.NewClient(client)
	bind.WaitMined(context.Background(), backend, tx)
}

func deploy(bytecode string) (common.Address, error) {
	cl, sk := getRealBackend()
	backend := ethclient.NewClient(cl)
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		return common.Address{}, err
	}
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		return common.Address{}, err
	}
	fmt.Printf("Nonce: %v\n", nonce)
	gp, _ := backend.SuggestGasPrice(context.Background())
	tx := types.NewContractCreation(nonce, common.Big0, 500000, gp.Mul(gp, common.Big2), common.Hex2Bytes(bytecode))
	signedTx, _ := types.SignTx(tx, types.NewLondonSigner(chainid), sk)
	if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
		return common.Address{}, err
	}
	return bind.WaitDeployed(context.Background(), backend, signedTx)
}

func execute(data []byte, gaslimit uint64) {
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
	tx := types.NewContractCreation(nonce, common.Big1, gaslimit, gp.Mul(gp, common.Big2), data)
	signedTx, _ := types.SignTx(tx, types.NewLondonSigner(chainid), sk)
	backend.SendTransaction(context.Background(), signedTx)
}

func randomBlobData() ([]byte, error) {
	val, err := rand.Int(rand.Reader, big.NewInt(maxDataPerBlob))
	if err != nil {
		return nil, err
	}
	size := int(val.Int64() * 3)
	data := make([]byte, size)
	n, err := rand.Read(data)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("could not create random blob data with size %d: %v", size, err)
	}
	return data, nil
}
