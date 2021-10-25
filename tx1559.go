package txfuzz

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// To1559Tx turns a normal transaction to a DynamicFeeTx.
func To1559Tx(tx *types.Transaction, chainID, tip, feeCap, gasPrice *big.Int, al *types.AccessList) *types.Transaction {
	v, r, s := tx.RawSignatureValues()
	list := tx.AccessList()
	if al != nil {
		list = *al
	}
	newTx := types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      tx.Nonce(),
		GasTipCap:  tip,
		GasFeeCap:  feeCap,
		Gas:        tx.Gas(),
		To:         tx.To(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: list,
		V:          v,
		R:          r,
		S:          s,
	}
	return types.NewTx(&newTx)
}
