package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

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

var (
	address = "http://127.0.0.1:8545"
)

func main() {
	// deployProxy
	addr, err := deployBlobProxy()
	if err != nil {
		panic(err)
	}

	// 4844 Tests

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

	// 4788 tests

	// Call addr
	contractAddr4788 := common.HexToAddress("0x01234") // TODO: update addr
	exec(contractAddr4788, Uint64ToHash(0).Bytes())

	t := time.Now().Unix()
	for i := 0; i < 15; i++ {
		exec(contractAddr4788, Uint64ToHash(uint64(t)-uint64(i)).Bytes())
	}

	// deploy4788Proxy
	addr, err = deploy4788Proxy()
	if err != nil {
		panic(err)
	}

	// Call to 4788 contract
	t = time.Now().Unix()
	for i := 0; i < 15; i++ {
		exec(addr, Uint64ToHash(uint64(t)-uint64(i)).Bytes())
	}

	// short or long calls
	for i := 0; i < 64; i++ {
		arr := make([]byte, i)
		exec(addr, arr)
	}

	// random calls
	for i := 0; i < 10; i++ {
		arr := make([]byte, 32)
		rand.Read(arr)
		exec(addr, arr)
	}
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
	signedTx, _ := types.SignTx(tx.Transaction, types.NewCancunSigner(chainid), sk)
	tx.Transaction = signedTx
	rlpData, err := tx.MarshalBinary()
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

func deployBlobProxy() (common.Address, error) {
	bytecode := "6080604052348015600f57600080fd5b5060ae80601d6000396000f3fe6080604052348015600f57600080fd5b506000366060600083838080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050905060008151602083016000f090505050915050805190602001f3fea2646970667358221220c23b98a79e6709c832ef1c90f5a3a7583ba88f759611d74a4d775dd22a02296364736f6c63430008120033"
	return deploy(bytecode)
}

/*
pragma solidity >=0.7.0 <0.9.0;
contract Call4788 {

    bytes val;

    fallback (bytes calldata _input) external returns (bytes memory _output) {
		address cntr = address(0xbEac00dDB15f3B6d645C48263dC93862413A222D);
		(bool success, bytes memory res) = cntr.call(_input);
        val = res;
    }
}
*/

func deploy4788Proxy() (common.Address, error) {
	bytecode := "608060405234801561001057600080fd5b5061046e806100206000396000f3fe608060405234801561001057600080fd5b506000366060600073beac00ddb15f3b6d645c48263dc93862413a222d90506000808273ffffffffffffffffffffffffffffffffffffffff1686866040516100599291906100fd565b6000604051808303816000865af19150503d8060008114610096576040519150601f19603f3d011682016040523d82523d6000602084013e61009b565b606091505b509150915080600090816100af9190610366565b50505050915050805190602001f35b600081905092915050565b82818337600083830152505050565b60006100e483856100be565b93506100f18385846100c9565b82840190509392505050565b600061010a8284866100d8565b91508190509392505050565b600081519050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b6000600282049050600182168061019757607f821691505b6020821081036101aa576101a9610150565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b6000600883026102127fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826101d5565b61021c86836101d5565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b600061026361025e61025984610234565b61023e565b610234565b9050919050565b6000819050919050565b61027d83610248565b6102916102898261026a565b8484546101e2565b825550505050565b600090565b6102a6610299565b6102b1818484610274565b505050565b5b818110156102d5576102ca60008261029e565b6001810190506102b7565b5050565b601f82111561031a576102eb816101b0565b6102f4846101c5565b81016020851015610303578190505b61031761030f856101c5565b8301826102b6565b50505b505050565b600082821c905092915050565b600061033d6000198460080261031f565b1980831691505092915050565b6000610356838361032c565b9150826002028217905092915050565b61036f82610116565b67ffffffffffffffff81111561038857610387610121565b5b610392825461017f565b61039d8282856102d9565b600060209050601f8311600181146103d057600084156103be578287015190505b6103c8858261034a565b865550610430565b601f1984166103de866101b0565b60005b82811015610406578489015182556001820191506020850194506020810190506103e1565b86831015610423578489015161041f601f89168261032c565b8355505b6001600288020188555050505b50505050505056fea26469706673582212201d7cb5ca006b5e6e8069b51068fdb7fb9e78fc5a45d7bc84d0a13748b777d44764736f6c63430008120033"
	return deploy(bytecode)
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

func Uint64ToHash(u uint64) common.Hash {
	var h common.Hash
	binary.BigEndian.PutUint64(h[24:], u)
	return h
}
