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
	if len(code) > 128 {
		code = code[:128]
	}
	switch f.Byte() {
	case 0:
		// Legacy contract creation
		return types.NewContractCreation(nonce, big.NewInt(0), gas, gasPrice, code), nil
	case 1:
		// Legacy transaction
		return types.NewTransaction(nonce, to, big.NewInt(0), gas, gasPrice, code), nil
	case 2:
		// AccessList contract creation
		return newALTx(nonce, nil, gas, chainID, gasPrice, big.NewInt(0), code, make(types.AccessList, 0)), nil
	case 3:
		// AccessList transaction
		return newALTx(nonce, &to, gas, chainID, gasPrice, big.NewInt(0), code, make(types.AccessList, 0)), nil
	case 4:
		// AccessList contract creation with AL
		tx := types.NewContractCreation(nonce, big.NewInt(0), gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		return newALTx(nonce, nil, gas, chainID, gasPrice, big.NewInt(0), code, *al), nil
	case 5:
		// AccessList transaction with AL
		tx := types.NewTransaction(nonce, to, big.NewInt(0), gas, gasPrice, code)
		al, err := CreateAccessList(rpc, tx, sender)
		if err != nil {
			return nil, err
		}
		return newALTx(nonce, &to, gas, chainID, gasPrice, big.NewInt(0), code, *al), nil
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
