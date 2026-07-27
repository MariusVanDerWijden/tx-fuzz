package spammer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// txBuilder creates the next transaction to send from an account.
type txBuilder func(backend *ethclient.Client, sender common.Address, nonce uint64) (*types.Transaction, error)

// sendTransactions sends config.N transactions from the account belonging to key
// and waits for the last one to be mined.
//
// A rejected transaction is an expected outcome here, not a failure: the node
// telling us a transaction is invalid is a result. Giving up on the whole round
// at the first rejection throws away the remaining transactions, so rejections
// are logged and the round continues. Only a round in which nothing at all got
// through is reported as an error, since that points at the setup rather than at
// the transactions.
func sendTransactions(config *Config, key *ecdsa.PrivateKey, build txBuilder) error {
	backend := ethclient.NewClient(config.backend)
	sender := crypto.PubkeyToAddress(key.PublicKey)
	chainID, err := backend.ChainID(context.Background())
	if err != nil {
		log.Warn("Could not get chainID, using default", "err", err)
		chainID = big.NewInt(0x01000666)
	}

	var (
		lastTx  *types.Transaction
		sent    int
		lastErr error
	)
	for i := uint64(0); i < config.N; i++ {
		nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(-1))
		if err != nil {
			return err
		}
		tx, err := build(backend, sender, nonce)
		if err != nil {
			log.Warn("Could not create transaction", "nonce", nonce, "err", err)
			lastErr = err
			continue
		}
		signedTx, err := txfuzz.SignTx(tx, chainID, key)
		if err != nil {
			return err
		}
		if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
			log.Warn("Could not submit transaction", "hash", signedTx.Hash(), "type", signedTx.Type(), "err", err)
			lastErr = err
			continue
		}
		lastTx = signedTx
		sent++
		time.Sleep(10 * time.Millisecond)
	}
	if sent == 0 {
		return fmt.Errorf("no transaction of %d was accepted, last error: %w", config.N, lastErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), TX_TIMEOUT)
	defer cancel()
	if _, err := bind.WaitMined(ctx, backend, lastTx); err != nil {
		fmt.Printf("Waiting for transactions to be mined failed: %v\n", err.Error())
	}
	return nil
}
