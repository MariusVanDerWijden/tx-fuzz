package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	maxDataPerBlob = 1 << 17 // 128Kb
)

var (
	address = "http://127.0.0.1:8545"
)

func main() {
	// deployProxy
	addr, err := deployProxy()
	if err != nil {
		panic(err)
	}

	// PUSH0, DATAHASH, PUSH0, DATAHASH, SSTORE
	exec(addr, []byte{0x5f, 0x49, 0x5f, 0x49, 0x55})

	var dataHashByteCode []byte
	for i := 0; i < 10; i++ {
		// PUSH1 i, DATAHASH, PUSH0, SSTORE
		dataHashByteCode = append(dataHashByteCode, []byte{0x60, byte(i), 0x49, 0x5f, 0x55}...)
	}
	exec(addr, dataHashByteCode)

	// PUSH1 0x01, NOT, DATAHASH, PUSH0, NOT, DATAHASH, SSTORE
	exec(addr, []byte{0x60, 0x01, 0x19, 0x49, 0x5f, 0x19, 0x49, 0x55})

	// PUSH1 0x255, PUSH1 0x01, SHL, DATAHASH, PUSH0, SSTORE
	exec(addr, []byte{0x60, 0xff, 0x1b, 0x49, 0x5f, 0x55})
}

func exec(addr common.Address, data []byte) {
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
	tip, _ := backend.SuggestGasTipCap(context.Background())
	blob, _ := randomBlobData()
	//nonce = nonce - 2
	tx := txfuzz.New4844Tx(nonce, &addr, 500000, chainid, tip.Mul(tip, common.Big1), gp.Mul(gp, common.Big1), common.Big0, data, big.NewInt(1000000), blob, make(types.AccessList, 0))
	signedTx, _ := types.SignTx(&tx.Transaction, types.NewCancunSigner(chainid), sk)
	tx.Transaction = *signedTx
	rlpData, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	if err := cl.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData)); err != nil {
		panic(err)
	}
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

func randomBlobData() ([]byte, error) {
	size := rand.Intn(maxDataPerBlob) * 3
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

/*
pragma solidity >=0.7.0 <0.9.0;
contract BlobProxy {
    fallback (bytes calldata _input) external returns (bytes memory _output) {
        bytes memory bytecode = _input;
        address addr;
        assembly {
            addr := create(0, add(bytecode, 0x20), mload(bytecode))
        }
    }
}
*/

func deployProxy() (common.Address, error) {
	bytecode := "6080604052348015600f57600080fd5b5060ae80601d6000396000f3fe6080604052348015600f57600080fd5b506000366060600083838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050905060008151602083016000f090505050915050805190602001f3fea2646970667358221220c23b98a79e6709c832ef1c90f5a3a7583ba88f759611d74a4d775dd22a02296364736f6c63430008120033"
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
