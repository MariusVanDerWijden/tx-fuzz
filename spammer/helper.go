package spammer

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const batchSize = 50

func SendTx(sk *ecdsa.PrivateKey, backend *ethclient.Client, to common.Address, value *big.Int) (*types.Transaction, error) {
	sender := crypto.PubkeyToAddress(sk.PublicKey)
	nonce, err := backend.NonceAt(context.Background(), sender, nil)
	if err != nil {
		fmt.Printf("Could not get pending nonce: %v", err)
	}
	return sendTxWithNonce(sk, backend, to, value, nonce)
}

func sendTxWithNonce(sk *ecdsa.PrivateKey, backend *ethclient.Client, to common.Address, value *big.Int, nonce uint64) (*types.Transaction, error) {
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	gp, _ := backend.SuggestGasPrice(context.Background())
	gas, _ := backend.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     crypto.PubkeyToAddress(sk.PublicKey),
		To:       &to,
		Gas:      30_000_000,
		GasPrice: gp,
		Value:    value,
		Data:     nil,
	})
	tx := types.NewTransaction(nonce, to, value, gas, gp.Mul(gp, big.NewInt(100)), nil)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainid), sk)
	return signedTx, backend.SendTransaction(context.Background(), signedTx)
}

func sendRecurringTx(sk *ecdsa.PrivateKey, backend *ethclient.Client, to common.Address, value *big.Int, numTxs uint64) (*types.Transaction, error) {
	sender := crypto.PubkeyToAddress(sk.PublicKey)
	nonce, err := backend.NonceAt(context.Background(), sender, nil)
	if err != nil {
		return nil, err
	}
	var (
		tx *types.Transaction
	)
	for i := 0; i < int(numTxs); i++ {
		tx, err = sendTxWithNonce(sk, backend, to, value, nonce+uint64(i))
	}
	return tx, err
}

func Unstuck(config *Config) error {
	if err := tryUnstuck(config, config.faucet); err != nil {
		return err
	}
	for _, key := range config.keys {
		if err := tryUnstuck(config, key); err != nil {
			return err
		}
	}
	return nil
}

func tryUnstuck(config *Config, sk *ecdsa.PrivateKey) error {
	var (
		client = ethclient.NewClient(config.backend)
		addr   = crypto.PubkeyToAddress(sk.PublicKey)
	)
	for i := 0; i < 100; i++ {
		noTx, err := isStuck(config, addr)
		if err != nil {
			return err
		}
		if noTx == 0 {
			return nil
		}

		// Self-transfer of 1 wei to unstuck
		if noTx > batchSize {
			noTx = batchSize
		}
		fmt.Println("Sending transaction to unstuck account")
		tx, err := sendRecurringTx(sk, client, addr, big.NewInt(1), noTx)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		if _, err := bind.WaitMined(ctx, client, tx); err != nil {
			return err
		}
	}
	fmt.Printf("Could not unstuck account %v after 100 tries\n", addr)
	return errors.New("unstuck timed out, please retry manually")
}

func isStuck(config *Config, account common.Address) (uint64, error) {
	client := ethclient.NewClient(config.backend)
	nonce, err := client.NonceAt(context.Background(), account, nil)
	if err != nil {
		return 0, err
	}

	pendingNonce, err := client.PendingNonceAt(context.Background(), account)
	if err != nil {
		return 0, err
	}

	if pendingNonce != nonce {
		fmt.Printf("Account %v is stuck: pendingNonce: %v currentNonce: %v, missing nonces: %v\n", account, pendingNonce, nonce, pendingNonce-nonce)
		return pendingNonce - nonce, nil
	}
	return 0, nil
}
