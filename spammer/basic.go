package spammer

import (
	"crypto/ecdsa"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const TX_TIMEOUT = 5 * time.Minute

// SendBasicTransactions sends transactions of every type that shares the plain
// transaction pool, which is everything but a blob transaction.
func SendBasicTransactions(config *Config, key *ecdsa.PrivateKey, f *filler.Filler) error {
	return sendTransactions(config, key, func(backend *ethclient.Client, sender common.Address, nonce uint64) (*types.Transaction, error) {
		return txfuzz.RandomTxWithOpts(config.backend, f, txfuzz.TxOpts{
			Sender:     sender,
			Nonce:      nonce,
			GasLimit:   config.gasLimit,
			AccessList: config.accessList,
			// Sign the EIP-7702 authorizations with the sender's own key, so
			// that set code transactions are part of the rotation.
			AuthKey: key,
		})
	})
}
