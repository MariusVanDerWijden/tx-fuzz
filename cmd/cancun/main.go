package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/MariusVanDerWijden/tx-fuzz/helper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/params"
)

func main() {
	fmt.Println("4788")
	test4788()
	fmt.Println("1153")
	test1153()
	fmt.Println("7516")
	test7516()
	fmt.Println("5656")
	test5656()
	fmt.Println("4844_prec")
	test4844_precompile()
	fmt.Println("4844")
	test4844()
}

func test7516() {
	// JUMPDEST, BLOBBASEFEE, POP, PUSH0, JUMP
	helper.Execute([]byte{0x5b, 0x4a, 0x50, 0x5f, 0x56}, 30_000_000)

	// BLOBBASEFEE, BLOBBASEFEE, SSTORE
	helper.Execute([]byte{0x4a, 0x4a, 0x50, 0x55}, 500_000)
}

func test5656() {
	// PUSH1, 0x20, PUSH0, PUSH0, MCOPY
	helper.Execute([]byte{0x60, 0x20, 0x5f, 0x5f, 0x5e}, 30_000_000)

	pushMaxVal := func() []byte {
		return []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	}

	sstore := func() []byte {
		return []byte{0x5a, 0x5a, 0x55}
	}

	// PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff, PUSH0, PUSH0, MCOPY
	helper.Execute(append(pushMaxVal(), 0x5f, 0x5f, 0x5e), 30_000_000)

	// PUSH0, PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff, PUSH0, MCOPY
	code := []byte{0x5f}
	code = append(code, pushMaxVal()...)
	code = append(code, []byte{0x5f, 0x5e}...)
	code = append(code, sstore()...)
	helper.Execute(code, 30_000_000)

	// PUSH0, PUSH0, PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff, MCOPY
	code = []byte{0x5f, 0x5f}
	code = append(code, pushMaxVal()...)
	code = append(code, []byte{0x5e}...)
	code = append(code, sstore()...)
	helper.Execute(code, 30_000_000)

	// PUSH0, PUSH1, 0xff, MSTORE, JUMPDEST, PUSH1, 0x20, PUSH0, PUSH0, MCOPY, PUSH1, 0x04 JUMP
	helper.Execute([]byte{0x5f, 0x60, 0xff, 0x52, 0x5b, 0x60, 0x20, 0x5f, 0x5f, 0x5e, 0x60, 0x04, 0x56}, 30_000_000)

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
	helper.Execute(code, 30_000_000)
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
	helper.Execute([]byte{0x5b, 0x5a, 0x5a, 0x5d, 0x5f, 0x56}, 30_000_000)

	// JUMPDEST, GAS, DUP1, DUP1, TSTORE, TLOAD, POP, PUSH0, JUMP
	helper.Execute([]byte{0x5b, 0x5a, 0x80, 0x80, 0x5d, 0x5c, 0x50, 0x5f, 0x56}, 30_000_000)

	// PUSH0, TLOAD, GAS, TLOAD, SSTORE
	helper.Execute([]byte{0x5f, 0x5c, 0x5a, 0x55}, 500_000)

	// PUSH32, 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff TLOAD DUP1, SSTORE
	helper.Execute([]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x5c, 0x80, 0x55}, 500_000)
}

func test4788() {
	// Call addr
	contractAddr4788 := common.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02")
	helper.Exec(contractAddr4788, Uint64ToHash(0).Bytes(), false)

	t := time.Now().Unix()
	for i := 0; i < 15; i++ {
		helper.Exec(contractAddr4788, Uint64ToHash(uint64(t)-uint64(i)).Bytes(), false)
	}

	// deploy4788Proxy
	addr, err := deploy4788Proxy()
	if err != nil {
		panic(err)
	}

	// Call to 4788 contract
	t = time.Now().Unix()
	for i := 0; i < 15; i++ {
		helper.Exec(addr, Uint64ToHash(uint64(t)-uint64(i)).Bytes(), false)
	}

	// short or long calls
	for i := 0; i < 64; i++ {
		arr := make([]byte, i)
		helper.Exec(addr, arr, false)
	}

	// random calls
	for i := 0; i < 10; i++ {
		arr := make([]byte, 32)
		rand.Read(arr)
		helper.Exec(addr, arr, false)
	}
}

func test4844() {
	addr, err := deployBlobProxy()
	if err != nil {
		panic(err)
	}

	// 4844 Tests

	// PUSH0, DATAHASH, PUSH0, DATAHASH, SSTORE
	helper.Exec(addr, []byte{0x5f, 0x49, 0x5f, 0x49, 0x55}, true)

	var dataHashByteCode []byte
	for i := 0; i < 10; i++ {
		// PUSH1 i, DATAHASH, PUSH0, SSTORE
		dataHashByteCode = append(dataHashByteCode, []byte{0x60, byte(i), 0x49, 0x5f, 0x55}...)
	}
	helper.Exec(addr, dataHashByteCode, true)

	// PUSH1 0x01, NOT, DATAHASH, PUSH0, NOT, DATAHASH, SSTORE
	helper.Exec(addr, []byte{0x60, 0x01, 0x19, 0x49, 0x5f, 0x19, 0x49, 0x55}, true)

	// PUSH1 0x255, PUSH1 0x01, SHL, DATAHASH, PUSH0, SSTORE
	tx := helper.Exec(addr, []byte{0x60, 0xff, 0x1b, 0x49, 0x5f, 0x55}, true)

	helper.Wait(tx)
}

func test4844_precompile() {
	addr, err := deployBlobCaller()
	if err != nil {
		panic(err)
	}

	// Test precompile without blobs
	staticTestInput := common.FromHex("01d18459b334ffe8e2226eef1db874fda6db2bdd9357268b39220af2d59464fb564c0a11a0f704f4fc3e8acfe0f8245f0ad1347b378fbf96e206da11a5d3630624d25032e67a7e6a4910df5834b8fe70e6bcfeeac0352434196bdf4b2485d5a1978a0d595c823c05947b1156175e72634a377808384256e9921ebf72181890be2d6b58d4a73a880541d1656875654806942307f266e636553e94006d11423f2688945ff3bdf515859eba1005c1a7708d620a94d91a1c0c285f9584e75ec2f82a")
	helper.Exec(addr, staticTestInput, false)

	invalidInput := make([]byte, len(staticTestInput))
	helper.Exec(addr, invalidInput, false)

	co, cl, pr, po, err := createPrecompileRandParams()
	if err != nil {
		panic(err)
	}
	validRandomInput := precompileParamsToBytes(co, cl, pr, po)
	tx := helper.Exec(addr, validRandomInput, false)

	helper.Wait(tx)

	// Test precompile with blobs
	helper.Exec(addr, staticTestInput, true)
	helper.Exec(addr, invalidInput, true)
	tx = helper.Exec(addr, validRandomInput, true)

	helper.Wait(tx)

	// Full block of verification
	for i := 0; i < 30_000_000/100_000; i++ {
		helper.Exec(addr, validRandomInput, false)
	}
}

func precompileParamsToBytes(commitment *kzg4844.Commitment, claim *kzg4844.Claim, proof *kzg4844.Proof, point *kzg4844.Point) []byte {
	bytes := make([]byte, 192)
	versionedHash := kZGToVersionedHash(commitment)
	copy(bytes[0:32], versionedHash[:])
	copy(bytes[32:64], point[:])
	copy(bytes[64:96], claim[:])
	copy(bytes[96:144], commitment[:])
	copy(bytes[144:192], proof[:])
	return bytes
}

func createPrecompileRandParams() (*kzg4844.Commitment, *kzg4844.Claim, *kzg4844.Proof, *kzg4844.Point, error) {
	random := make([]byte, 131072)
	rand.Read(random[:])
	blob := encodeBlobs(random)[0]
	commitment, err := kzg4844.BlobToCommitment(&blob)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var point kzg4844.Point
	rand.Read(point[:])
	point[0] = 0 // point needs to be < modulus
	proof, claim, err := kzg4844.ComputeProof(&blob, point)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return &commitment, &claim, &proof, &point, nil
}

func encodeBlobs(data []byte) []kzg4844.Blob {
	blobs := []kzg4844.Blob{{}}
	blobIndex := 0
	fieldIndex := -1
	for i := 0; i < len(data); i += 31 {
		fieldIndex++
		if fieldIndex == params.BlobTxFieldElementsPerBlob {
			blobs = append(blobs, kzg4844.Blob{})
			blobIndex++
			fieldIndex = 0
		}
		max := i + 31
		if max > len(data) {
			max = len(data)
		}
		copy(blobs[blobIndex][fieldIndex*32+1:], data[i:max])
	}
	return blobs
}

// kZGToVersionedHash implements kzg_to_versioned_hash from EIP-4844
func kZGToVersionedHash(kzg *kzg4844.Commitment) common.Hash {
	h := sha256.Sum256(kzg[:])
	h[0] = 0x01

	return h
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
	return helper.Deploy(bytecode)
}

/*
pragma solidity >=0.7.0 <0.9.0;

	contract BlobCaller {
	    bool _ok;
	    bytes out;

	    fallback (bytes calldata _input) external returns (bytes memory _output) {
	        address precompile = address(0x0A);
	        (bool ok, bytes memory output) = precompile.call{gas: 50000}(_input);
	        _output = output;
	        // Store return values to trigger sstore
	        _ok = ok;
	        out = output;
	    }
	}
*/
func deployBlobCaller() (common.Address, error) {
	bytecode := "608060405234801561001057600080fd5b5061047b806100206000396000f3fe608060405234801561001057600080fd5b5060003660606000600a90506000808273ffffffffffffffffffffffffffffffffffffffff1661c350878760405161004992919061010a565b60006040518083038160008787f1925050503d8060008114610087576040519150601f19603f3d011682016040523d82523d6000602084013e61008c565b606091505b5091509150809350816000806101000a81548160ff02191690831515021790555080600190816100bc9190610373565b50505050915050805190602001f35b600081905092915050565b82818337600083830152505050565b60006100f183856100cb565b93506100fe8385846100d6565b82840190509392505050565b60006101178284866100e5565b91508190509392505050565b600081519050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806101a457607f821691505b6020821081036101b7576101b661015d565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b60006008830261021f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826101e2565b61022986836101e2565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b600061027061026b61026684610241565b61024b565b610241565b9050919050565b6000819050919050565b61028a83610255565b61029e61029682610277565b8484546101ef565b825550505050565b600090565b6102b36102a6565b6102be818484610281565b505050565b5b818110156102e2576102d76000826102ab565b6001810190506102c4565b5050565b601f821115610327576102f8816101bd565b610301846101d2565b81016020851015610310578190505b61032461031c856101d2565b8301826102c3565b50505b505050565b600082821c905092915050565b600061034a6000198460080261032c565b1980831691505092915050565b60006103638383610339565b9150826002028217905092915050565b61037c82610123565b67ffffffffffffffff8111156103955761039461012e565b5b61039f825461018c565b6103aa8282856102e6565b600060209050601f8311600181146103dd57600084156103cb578287015190505b6103d58582610357565b86555061043d565b601f1984166103eb866101bd565b60005b82811015610413578489015182556001820191506020850194506020810190506103ee565b86831015610430578489015161042c601f891682610339565b8355505b6001600288020188555050505b50505050505056fea264697066735822122089d7332a134ee7e7d76876ef5f4e74d939f9d9d9f3344e6afb518c96fff0b63164736f6c63430008120033"
	return helper.Deploy(bytecode)
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
	return helper.Deploy(bytecode)
}

func Uint64ToHash(u uint64) common.Hash {
	var h common.Hash
	binary.BigEndian.PutUint64(h[24:], u)
	return h
}
