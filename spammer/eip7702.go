package spammer

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"math/rand"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Send7702Transactions sends EIP-7702 set code transactions with fuzzed
// authorization lists.
func Send7702Transactions(config *Config, key *ecdsa.PrivateKey, f *filler.Filler) error {
	return sendTransactions(config, key, func(backend *ethclient.Client, sender common.Address, nonce uint64) (*types.Transaction, error) {
		// Half the time authorize the sender itself, half the time another
		// account we hold a key for. A self-sponsored authorization takes a
		// different path through the state transition than a sponsored one.
		authorizer, authNonce := key, nonce
		if f.Bool() {
			authorizer = config.keys[rand.Intn(len(config.keys))]
			var err error
			authNonce, err = backend.NonceAt(context.Background(), crypto.PubkeyToAddress(authorizer.PublicKey), big.NewInt(-1))
			if err != nil {
				return nil, err
			}
		}
		return txfuzz.RandomSetCodeTxWithOpts(config.backend, f, txfuzz.TxOpts{
			Sender:     sender,
			Nonce:      nonce,
			GasLimit:   config.gasLimit,
			AccessList: config.accessList,
			AuthKey:    authorizer,
			AuthNonce:  authNonce,
		})
	})
}
