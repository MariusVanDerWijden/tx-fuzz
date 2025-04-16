package helper

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum"
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

func Exec(addr common.Address, data []byte, blobs bool) *types.Transaction {
	_, sk := GetRealBackend()
	return ExecWithSK(sk, addr, data, blobs)
}

func ExecWithSK(sk *ecdsa.PrivateKey, addr common.Address, data []byte, blobs bool) *types.Transaction {
	cl, _ := GetRealBackend()
	backend := ethclient.NewClient(cl)
	sender := crypto.PubkeyToAddress(sk.PublicKey)
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
		blob, err := RandomBlobData()
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

func ExecAuth(addr common.Address, data []byte, authList []types.SetCodeAuthorization) *types.Transaction {
	cl, sk := GetRealBackend()
	backend := ethclient.NewClient(cl)
	sender := crypto.PubkeyToAddress(sk.PublicKey)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		panic(err)
	}
	return ExecAuthWithNonce(addr, nonce, data, authList)
}

func ExecAuthWithNonce(addr common.Address, nonce uint64, data []byte, authList []types.SetCodeAuthorization) *types.Transaction {
	cl, sk := GetRealBackend()
	backend := ethclient.NewClient(cl)
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
	var rlpData []byte
	var _tx *types.Transaction
	gasLimit := uint64(5_000_000)
	if authList == nil {
		buf := make([]byte, 1024)
		rand.Read(buf)
		aList, err := txfuzz.RandomAuthList(filler.NewFiller(buf), sk)
		if err != nil {
			panic(err)
		}
		authList = aList
	}
	tx := txfuzz.New7702Tx(nonce, addr, gasLimit, chainid, tip.Mul(tip, big.NewInt(100)), gp.Mul(gp, big.NewInt(100)), common.Big0, data, big.NewInt(1_000_000), make(types.AccessList, 0), authList)
	signedTx, _ := types.SignTx(tx, types.NewPragueSigner(chainid), sk)
	rlpData, err = signedTx.MarshalBinary()
	if err != nil {
		panic(err)
	}
	_tx = signedTx
	if err := cl.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData)); err != nil {
		fmt.Println(err)
	}
	return _tx
}

func GetRealBackend() (*rpc.Client, *ecdsa.PrivateKey) {
	// eth.sendTransaction({from:personal.listAccounts[0], to:"0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D", value: "100000000000000"})

	sk := crypto.ToECDSAUnsafe(common.FromHex(txfuzz.SK))
	if crypto.PubkeyToAddress(sk.PublicKey).Hex() != txfuzz.ADDR {
		panic(fmt.Sprintf("wrong address want %s got %s", crypto.PubkeyToAddress(sk.PublicKey).Hex(), txfuzz.ADDR))
	}
	cl, err := rpc.Dial(txfuzz.RPC)
	if err != nil {
		panic(err)
	}
	return cl, sk
}

func Wait(tx *types.Transaction) {
	client, _ := GetRealBackend()
	backend := ethclient.NewClient(client)
	bind.WaitMined(context.Background(), backend, tx)
}

func ChainID() *big.Int {
	cl, _ := GetRealBackend()
	backend := ethclient.NewClient(cl)
	id, err := backend.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	return id
}

func Nonce(addr common.Address) uint64 {
	cl, _ := GetRealBackend()
	backend := ethclient.NewClient(cl)
	nonce, err := backend.NonceAt(context.Background(), addr, nil)
	if err != nil {
		panic(err)
	}
	return nonce
}

func Deploy(bytecode string) (common.Address, error) {
	cl, sk := GetRealBackend()
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
	tx := types.NewContractCreation(nonce, common.Big0, 5_000_000, gp.Mul(gp, common.Big2), common.Hex2Bytes(bytecode))
	signedTx, _ := types.SignTx(tx, types.NewCancunSigner(chainid), sk)
	if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
		return common.Address{}, err
	}
	return bind.WaitDeployed(context.Background(), backend, signedTx)
}

func Execute(data []byte, gaslimit uint64) error {
	cl, sk := GetRealBackend()
	backend := ethclient.NewClient(cl)
	sender := crypto.PubkeyToAddress(sk.PublicKey)
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
	return backend.SendTransaction(context.Background(), signedTx)
}

func RandomBlobData() ([]byte, error) {
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
