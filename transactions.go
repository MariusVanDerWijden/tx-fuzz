package txfuzz

import (
	"context"
	"errors"
	"math/big"
	"math/rand"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/MariusVanDerWijden/FuzzyVM/generator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
	"github.com/protolambda/ztyp/view"
)

// RandomCode creates a random byte code from the passed filler.
func RandomCode(f *filler.Filler) []byte {
	_, code := generator.GenerateProgram(f)
	return code
}

// RandomTx creates a random transaction.
func RandomTx(f *filler.Filler) (*types.Transaction, error) {
	nonce := uint64(rand.Int63())
	gasPrice := big.NewInt(rand.Int63())
	chainID := big.NewInt(rand.Int63())
	return RandomValidTx(nil, f, common.Address{}, nonce, gasPrice, chainID, false)
}

// RandomValidTx creates a random valid transaction.
// It does not mean that the transaction will succeed, but that it is well-formed.
// If gasPrice is not set, we will try to get it from the rpc
// If chainID is not set, we will try to get it from the rpc
func RandomValidTx(rpc *rpc.Client, f *filler.Filler, sender common.Address, nonce uint64, gasPrice, chainID *big.Int, al bool) (*types.Transaction, error) {
	// Set fields if non-nil
	if rpc != nil {
		client := ethclient.NewClient(rpc)
		var err error
		if gasPrice == nil {
			gasPrice, err = client.SuggestGasPrice(context.Background())
			if err != nil {
				gasPrice = big.NewInt(1)
			}
		}
		if chainID == nil {
			chainID, err = client.ChainID(context.Background())
			if err != nil {
				chainID = big.NewInt(1)
			}
		}
	}
	gas := uint64(100000)
	to := randomAddress()
	code := RandomCode(f)
	value := big.NewInt(0)
	if len(code) > 128 {
		code = code[:128]
	}
	mod := 12
	if al {
		mod = 6
	}
	switch f.Byte() % byte(mod) {
	case 0:
		// Legacy contract creation
		return types.NewContractCreation(nonce, value, gas, gasPrice, code), nil
	case 1:
		// Legacy transaction
		return types.NewTransaction(nonce, to, value, gas, gasPrice, code), nil
	case 2:
		// AccessList contract creation
		return newALTx(nonce, nil, gas, chainID, gasPrice, value, code, make(types.AccessList, 0)), nil
	case 3:
		// AccessList transaction
		return newALTx(nonce, &to, gas, chainID, gasPrice, value, code, make(types.AccessList, 0)), nil

	case 4:
		// 1559 contract creation
		tip, feecap, err := getCaps(rpc, gasPrice)
		if err != nil {
			return nil, err
		}
		return new1559Tx(nonce, nil, gas, chainID, tip, feecap, value, code, make(types.AccessList, 0)), nil
	case 5:
		// 1559 transaction
		tip, feecap, err := getCaps(rpc, gasPrice)
		if err != nil {
			return nil, err
		}
		return new1559Tx(nonce, &to, gas, chainID, tip, feecap, value, code, make(types.AccessList, 0)), nil
	case 6:
		// AccessList contract creation with AL
		tx := types.NewContractCreation(nonce, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		return newALTx(nonce, nil, gas, chainID, gasPrice, value, code, *al), nil
	case 7:
		// AccessList transaction with AL
		tx := types.NewTransaction(nonce, to, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		return newALTx(nonce, &to, gas, chainID, gasPrice, value, code, *al), nil
	case 8:
		// 1559 contract creation with AL
		tx := types.NewContractCreation(nonce, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		tip, feecap, err := getCaps(rpc, gasPrice)
		if err != nil {
			return nil, err
		}
		return new1559Tx(nonce, nil, gas, chainID, tip, feecap, value, code, *al), nil
	case 9:
		// 1559 tx with AL
		tx := types.NewTransaction(nonce, to, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		tip, feecap, err := getCaps(rpc, gasPrice)
		if err != nil {
			return nil, err
		}
		return new1559Tx(nonce, &to, gas, chainID, tip, feecap, value, code, *al), nil
	case 10:
		// 4844 transaction with AL
		tx := types.NewTransaction(nonce, to, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		tip, feecap, err := getCaps(rpc, gasPrice)
		if err != nil {
			return nil, err
		}
		data, err := randomBlobData()
		if err != nil {
			return nil, err
		}
		return new4844Tx(nonce, &to, gas, chainID, tip, feecap, value, code, data, *al), nil
	case 11:
		// 4844 transaction without AL
		tip, feecap, err := getCaps(rpc, gasPrice)
		if err != nil {
			return nil, err
		}
		data, err := randomBlobData()
		if err != nil {
			return nil, err
		}
		return new4844Tx(nonce, &to, gas, chainID, tip, feecap, value, code, data, make(types.AccessList, 0)), nil
	}
	return nil, errors.New("asdf")
}

func newALTx(nonce uint64, to *common.Address, gasLimit uint64, chainID, gasPrice, value *big.Int, code []byte, al types.AccessList) *types.Transaction {
	return types.NewTx(&types.AccessListTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasPrice:   gasPrice,
		Gas:        gasLimit,
		To:         to,
		Value:      value,
		Data:       code,
		AccessList: al,
	})
}

func new1559Tx(nonce uint64, to *common.Address, gasLimit uint64, chainID, tip, feeCap, value *big.Int, code []byte, al types.AccessList) *types.Transaction {
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  tip,
		GasFeeCap:  feeCap,
		Gas:        gasLimit,
		To:         to,
		Value:      value,
		Data:       code,
		AccessList: al,
	})
}

func new4844Tx(nonce uint64, to *common.Address, gasLimit uint64, chainID, tip, feeCap, value *big.Int, code, blobData []byte, al types.AccessList) *types.Transaction {
	cp, ok := uint256.FromBig(feeCap)
	if !ok {
		panic("fee cap not big int")
	}
	tp, ok := uint256.FromBig(tip)
	if !ok {
		panic("tip not big int")
	}
	val, ok := uint256.FromBig(value)
	if !ok {
		panic("value not big int")
	}
	chID, ok := uint256.FromBig(chainID)
	if !ok {
		panic("chainID not big int")
	}
	blobs := encodeBlobs(blobData)
	commits, versionedHashes, aggProof, err := blobs.ComputeCommitmentsAndAggregatedProof()
	if err != nil {
		panic(err)
	}
	txData := types.SignedBlobTx{
		Message: types.BlobTxMessage{
			ChainID:             view.Uint256View(*chID),
			Nonce:               view.Uint64View(nonce),
			GasTipCap:           view.Uint256View(*tp),
			GasFeeCap:           view.Uint256View(*cp),
			Gas:                 view.Uint64View(gasLimit),
			To:                  types.AddressOptionalSSZ{Address: (*types.AddressSSZ)(to)},
			Value:               view.Uint256View(*val),
			Data:                types.TxDataView(code),
			AccessList:          types.AccessListView(al),
			BlobVersionedHashes: versionedHashes,
			MaxFeePerDataGas:    view.Uint256View(*cp), // Use the same fee cap as gas for now.
		},
	}
	wrapData := types.BlobTxWrapData{
		BlobKzgs:           commits,
		Blobs:              blobs,
		KzgAggregatedProof: aggProof,
	}
	return types.NewTx(&txData, types.WithTxWrapData(&wrapData))
}

func getCaps(rpc *rpc.Client, defaultGasPrice *big.Int) (*big.Int, *big.Int, error) {
	if rpc == nil {
		tip := new(big.Int).Mul(big.NewInt(1), big.NewInt(params.GWei))
		if defaultGasPrice.Cmp(tip) >= 0 {
			feeCap := new(big.Int).Sub(defaultGasPrice, tip)
			return tip, feeCap, nil
		}
		return big.NewInt(0), defaultGasPrice, nil
	}
	client := ethclient.NewClient(rpc)
	tip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, nil, err
	}
	feeCap, err := client.SuggestGasPrice(context.Background())
	return tip, feeCap, err
}

func encodeBlobs(data []byte) types.Blobs {
	blobs := []types.Blob{{}}
	blobIndex := 0
	fieldIndex := -1
	for i := 0; i < len(data); i += 31 {
		fieldIndex++
		if fieldIndex == params.FieldElementsPerBlob {
			blobs = append(blobs, types.Blob{})
			blobIndex++
			fieldIndex = 0
		}
		max := i + 31
		if max > len(data) {
			max = len(data)
		}
		copy(blobs[blobIndex][fieldIndex][:], data[i:max])
	}
	return blobs
}
