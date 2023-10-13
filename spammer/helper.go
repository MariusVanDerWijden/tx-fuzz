package spammer

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func SendTx(sk *ecdsa.PrivateKey, backend *ethclient.Client, to common.Address, value *big.Int) (*types.Transaction, error) {
	sender := common.HexToAddress(txfuzz.ADDR)
	nonce, err := backend.PendingNonceAt(context.Background(), sender)
	if err != nil {
		fmt.Printf("Could not get pending nonce: %v", err)
	}
	fmt.Printf("Nonce: %v\n", nonce)
	chainid, err := backend.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	gp, _ := backend.SuggestGasPrice(context.Background())
	tx := types.NewTransaction(nonce, to, value, 500000, gp, nil)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainid), sk)
	return signedTx, backend.SendTransaction(context.Background(), signedTx)
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

		fmt.Println("Sending transaction to unstuck account")
		// Self-transfer of 1 wei to unstuck
		tx, err := SendTx(sk, client, addr, big.NewInt(1))
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Second)
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
		fmt.Printf("Account %v is stuck: pendingNonce: %v currentNonce: %v\n", account, pendingNonce, nonce)
		return pendingNonce - nonce, nil
	}
	return 0, nil
}
