package txfuzz

import (
	"context"
	"errors"
	"math/big"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/MariusVanDerWijden/FuzzyVM/generator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// RandomCode creates a random byte code from the passed filler.
func RandomCode(f *filler.Filler) []byte {
	_, code := generator.GenerateProgram(f)
	return code
}

// RandomTx creates a random transaction.
func RandomTx(f *filler.Filler) {

}

// RandomValidTx creates a random valid transaction.
// It does not mean that the transaction will succeed, but that it is well-formed.
// If gasPrice is not set, we will try to get it from the rpc
// If chainID is not set, we will try to get it from the rpc
func RandomValidTx(rpc *rpc.Client, f *filler.Filler, sender common.Address, nonce uint64, gasPrice, chainID *big.Int) (*types.Transaction, error) {
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
	gas := uint64(300000)
	to := randomAddress()
	code := RandomCode(f)
	value := big.NewInt(0)
	if len(code) > 128 {
		code = code[:128]
	}
	switch f.Byte() % 10 {
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
		// AccessList contract creation with AL
		tx := types.NewContractCreation(nonce, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		return newALTx(nonce, nil, gas, chainID, gasPrice, value, code, *al), nil
	case 5:
		// AccessList transaction with AL
		tx := types.NewTransaction(nonce, to, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		return newALTx(nonce, &to, gas, chainID, gasPrice, value, code, *al), nil
	case 6:
		// 1559 contract creation
		tip, feecap, err := getCaps(rpc)
		if err != nil {
			return nil, err
		}
		return new1559Tx(nonce, nil, gas, chainID, tip, feecap, value, code, make(types.AccessList, 0)), nil
	case 7:
		// 1559 transaction
		tip, feecap, err := getCaps(rpc)
		if err != nil {
			return nil, err
		}
		return new1559Tx(nonce, &to, gas, chainID, tip, feecap, value, code, make(types.AccessList, 0)), nil
	case 8:
		// 1559 contract creation with AL
		tx := types.NewContractCreation(nonce, value, gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		tip, feecap, err := getCaps(rpc)
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
		tip, feecap, err := getCaps(rpc)
		if err != nil {
			return nil, err
		}
		return new1559Tx(nonce, &to, gas, chainID, tip, feecap, value, code, *al), nil
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

func getCaps(rpc *rpc.Client) (*big.Int, *big.Int, error) {
	client := ethclient.NewClient(rpc)
	tip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, nil, err
	}
	feeCap, err := client.SuggestGasPrice(context.Background())
	return tip, feeCap, err
}
