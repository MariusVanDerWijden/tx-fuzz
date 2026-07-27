package spammer

import (
	"crypto/ecdsa"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SendBlobTransactions sends blob transactions. They live in a pool of their
// own, which is why they are not part of the basic rotation.
func SendBlobTransactions(config *Config, key *ecdsa.PrivateKey, f *filler.Filler) error {
	return sendTransactions(config, key, func(backend *ethclient.Client, sender common.Address, nonce uint64) (*types.Transaction, error) {
		return txfuzz.RandomBlobTxWithOpts(config.backend, f, txfuzz.TxOpts{
			Sender:         sender,
			Nonce:          nonce,
			GasLimit:       config.gasLimit,
			AccessList:     config.accessList,
			SidecarVersion: config.SidecarVersion,
		})
	})
}
