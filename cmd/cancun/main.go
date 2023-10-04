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
	test4844()
	//test4788()
	//test1153()
	//test7516()
	//test5656()
}

func test7516() {
	// JUMPDEST, BLOBBASEFEE, POP, PUSH0, JUMP
	execute([]byte{0x5b, 0x4a, 0x50, 0x5f, 0x56}, 30_000_000)

	// BLOBBASEFEE, BLOBBASEFEE, SSTORE
	execute([]byte{0x4a, 0x4a, 0x50, 0x55}, 500_000)
}

func test5656() {
	// PUSH1, 0x20, PUSH0, PUSH0, MCOPY
	execute([]byte{0x60, 0x20, 0x5f, 0x5f, 0x5e}, 30_000_000)

	pushMaxVal := func() []byte {
		return []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	}

	sstore := func() []byte {
		return []byte{0x5a, 0x5a, 0x55}
	}

	// PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff, PUSH0, PUSH0, MCOPY
	execute(append(pushMaxVal(), 0x5f, 0x5f, 0x5e), 30_000_000)

	// PUSH0, PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff, PUSH0, MCOPY
	code := []byte{0x5f}
	code = append(code, pushMaxVal()...)
	code = append(code, []byte{0x5f, 0x5e}...)
	code = append(code, sstore()...)
	execute(code, 30_000_000)

	// PUSH0, PUSH0, PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff, MCOPY
	code = []byte{0x5f, 0x5f}
	code = append(code, pushMaxVal()...)
	code = append(code, []byte{0x5e}...)
	code = append(code, sstore()...)
	execute(code, 30_000_000)

	// PUSH0, PUSH1, 0xff, MSTORE, JUMPDEST, PUSH1, 0x20, PUSH0, PUSH0, MCOPY, PUSH1, 0x04 JUMP
	execute([]byte{0x5f, 0x60, 0xff, 0x52, 0x5b, 0x60, 0x20, 0x5f, 0x5f, 0x5e, 0x60, 0x04, 0x56}, 30_000_000)

	test5656_memBuster()
}

func test5656_memBuster() {
	// Create N * 32 bytes of memory, Then copies them n times
	N := 1024
	n := 1024
	code := []byte{}
	for i := 0; i < n; i++ {
		code = append(code, pushSize(i)...)        // PUSH4 i
		code = append(code, []byte{0x60, 0xff}...) // PUSH1 0xff
		code = append(code, 0x52)                  // MSTORE
	}
	size := n
	for i := 0; i < N; i++ {
		code = append(code, pushSize(size)...) // PUSH4 size (length)
		code = append(code, 0x5f)              // PUSH0 (offset)
		code = append(code, pushSize(size)...) // PUSH4 size (dst)
		code = append(code, 0x5e)              // MCOPY
		size = size * 2
		if size > 0x2000 {
			break
		}
	}
	code = append(code, pushSize(size)...) // PUSH4 size (dst offset)
	jumpdest := len(code)
	code = append(code, 0x5b)              // JUMPDEST
	code = append(code, pushSize(size)...) // PUSH4 size (length)
	code = append(code, 0x5f)              // PUSH0 (offset)
	code = append(code, 0x82)              // DUP3 (dst)
	code = append(code, 0x5e)              // MCOPY
	// Add size to dst offset
	code = append(code, pushSize(size)...) // PUSH4 size
	code = append(code, 0x01)              // ADD
	code = append(code, pushSize(jumpdest)...)
	code = append(code, 0x56)

	// PUSH0, PUSH1, 0xff, MSTORE, JUMPDEST, PUSH1, 0x20, PUSH0, PUSH0, MCOPY, PUSH1, 0x04 JUMP
	execute(code, 30_000_000)
}

func pushSize(size int) []byte {
	code := []byte{0x63}
	sizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeArr, uint32(size))
	code = append(code, sizeArr...)
	return code
}

func test1153() {
	// JUMPDEST, GAS, GAS, TSTORE, PUSH0, JUMP
	execute([]byte{0x5b, 0x5a, 0x5a, 0x5d, 0x5f, 0x56}, 30_000_000)

	// JUMPDEST, GAS, DUP1, DUP1, TSTORE, TLOAD, POP, PUSH0, JUMP
	execute([]byte{0x5b, 0x5a, 0x80, 0x80, 0x5d, 0x5c, 0x50, 0x5f, 0x56}, 30_000_000)

	// PUSH0, TLOAD, GAS, TLOAD, SSTORE
	execute([]byte{0x5f, 0x5c, 0x5a, 0x55}, 500_000)

	// PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff TLOAD DUP1, SSTORE
	execute([]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x5c, 0x80, 0x55}, 500_000)
}

func test4788() {
	// Call addr
	contractAddr4788 := common.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02")
	exec(contractAddr4788, Uint64ToHash(0).Bytes(), false)

	t := time.Now().Unix()
	for i := 0; i < 15; i++ {
		exec(contractAddr4788, Uint64ToHash(uint64(t)-uint64(i)).Bytes(), false)
	}

	// deploy4788Proxy
	addr, err := deploy4788Proxy()
	if err != nil {
		panic(err)
	}

	// Call to 4788 contract
	t = time.Now().Unix()
	for i := 0; i < 15; i++ {
		exec(addr, Uint64ToHash(uint64(t)-uint64(i)).Bytes(), false)
	}

	// short or long calls
	for i := 0; i < 64; i++ {
		arr := make([]byte, i)
		exec(addr, arr, false)
	}

	// random calls
	for i := 0; i < 10; i++ {
		arr := make([]byte, 32)
		rand.Read(arr)
		exec(addr, arr, false)
	}
}

func test4844() {
	// deployProxy
	addr, err := deployBlobProxy()
	if err != nil {
		panic(err)
	}

	// 4844 Tests

	// PUSH0, DATAHASH, PUSH0, DATAHASH, SSTORE
	exec(addr, []byte{0x5f, 0x49, 0x5f, 0x49, 0x55}, true)

	var dataHashByteCode []byte
	for i := 0; i < 10; i++ {
		// PUSH1 i, DATAHASH, PUSH0, SSTORE
		dataHashByteCode = append(dataHashByteCode, []byte{0x60, byte(i), 0x49, 0x5f, 0x55}...)
	}
	exec(addr, dataHashByteCode, true)

	// PUSH1 0x01, NOT, DATAHASH, PUSH0, NOT, DATAHASH, SSTORE
	exec(addr, []byte{0x60, 0x01, 0x19, 0x49, 0x5f, 0x19, 0x49, 0x55}, true)

	// PUSH1 0x255, PUSH1 0x01, SHL, DATAHASH, PUSH0, SSTORE
	exec(addr, []byte{0x60, 0xff, 0x1b, 0x49, 0x5f, 0x55}, true)
}

func exec(addr common.Address, data []byte, blobs bool) {
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
	if blobs {
		blob, err := randomBlobData()
		if err != nil {
			panic(err)
		}
		//nonce = nonce - 2
		tx := txfuzz.New4844Tx(nonce, &addr, 500000, chainid, tip.Mul(tip, common.Big1), gp.Mul(gp, common.Big1), common.Big0, data, big.NewInt(1000000), blob, make(types.AccessList, 0))
		signedTx, _ := types.SignTx(tx.Transaction, types.NewCancunSigner(chainid), sk)
		tx.Transaction = signedTx
		rlpData, err = tx.MarshalBinary()
		if err != nil {
			panic(err)
		}
	} else {
		tx := types.NewTx(&types.DynamicFeeTx{ChainID: chainid, Nonce: nonce, GasTipCap: tip, GasFeeCap: gp, Gas: 500000, To: &addr, Data: data})
		signedTx, _ := types.SignTx(tx, types.NewCancunSigner(chainid), sk)
		rlpData, err = signedTx.MarshalBinary()
		if err != nil {
			panic(err)
		}
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
    bool suc;

    fallback (bytes calldata _input) external returns (bytes memory _output) {
		address cntr = address(0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02);
		(bool success, bytes memory res) = cntr.call(_input);
        val = res;
        suc = success;
        return res;
    }
}
*/

func deploy4788Proxy() (common.Address, error) {
	bytecode := "608060405234801561001057600080fd5b5061048a806100206000396000f3fe608060405234801561001057600080fd5b5060003660606000720f3df6d732807ef1319fb7b8bb8522d0beac0290506000808273ffffffffffffffffffffffffffffffffffffffff168686604051610058929190610119565b6000604051808303816000865af19150503d8060008114610095576040519150601f19603f3d011682016040523d82523d6000602084013e61009a565b606091505b509150915080600090816100ae9190610382565b5081600160006101000a81548160ff021916908315150217905550809350505050915050805190602001f35b600081905092915050565b82818337600083830152505050565b600061010083856100da565b935061010d8385846100e5565b82840190509392505050565b60006101268284866100f4565b91508190509392505050565b600081519050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806101b357607f821691505b6020821081036101c6576101c561016c565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b60006008830261022e7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826101f1565b61023886836101f1565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b600061027f61027a61027584610250565b61025a565b610250565b9050919050565b6000819050919050565b61029983610264565b6102ad6102a582610286565b8484546101fe565b825550505050565b600090565b6102c26102b5565b6102cd818484610290565b505050565b5b818110156102f1576102e66000826102ba565b6001810190506102d3565b5050565b601f82111561033657610307816101cc565b610310846101e1565b8101602085101561031f578190505b61033361032b856101e1565b8301826102d2565b50505b505050565b600082821c905092915050565b60006103596000198460080261033b565b1980831691505092915050565b60006103728383610348565b9150826002028217905092915050565b61038b82610132565b67ffffffffffffffff8111156103a4576103a361013d565b5b6103ae825461019b565b6103b98282856102f5565b600060209050601f8311600181146103ec57600084156103da578287015190505b6103e48582610366565b86555061044c565b601f1984166103fa866101cc565b60005b82811015610422578489015182556001820191506020850194506020810190506103fd565b8683101561043f578489015161043b601f891682610348565b8355505b6001600288020188555050505b50505050505056fea2646970667358221220d3505d93c72ff246e512c416145e275fe92925a05a9953337b1add26b509ec7764736f6c63430008120033"
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

func Uint64ToHash(u uint64) common.Hash {
	var h common.Hash
	binary.BigEndian.PutUint64(h[24:], u)
	return h
}
