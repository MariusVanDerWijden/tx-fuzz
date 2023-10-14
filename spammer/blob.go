package spammer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

func SendBlobTransactions(config *Config, key *ecdsa.PrivateKey, f *filler.Filler) error {
	backend := ethclient.NewClient(config.backend)
	sender := crypto.PubkeyToAddress(key.PublicKey)
	chainID, err := backend.ChainID(context.Background())
	if err != nil {
		log.Warn("Could not get chainID, using default")
		chainID = big.NewInt(0x01000666)
	}

	var lastTx *types.Transaction
	for i := uint64(0); i < config.N; i++ {
		nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(-1))
		if err != nil {
			return err
		}
		tx, err := txfuzz.RandomBlobTx(config.backend, f, sender, nonce, nil, nil, config.accessList)
		if err != nil {
			log.Warn("Could not create valid tx: %v", nonce)
			return err
		}
		signedTx, err := types.SignTx(tx.Transaction, types.NewCancunSigner(chainID), key)
		if err != nil {
			return err
		}
		tx.Transaction = signedTx
		rlpData, err := tx.MarshalBinary()
		if err != nil {
			return err
		}
		if err := config.backend.CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData)); err != nil {
			if strings.Contains(err.Error(), "account limit exceeded") {
				// Back off for a bit if we send a lot of transactions at once
				time.Sleep(1 * time.Minute)
				continue
			} else {
				return err
			}
		}
		lastTx = signedTx
		time.Sleep(10 * time.Millisecond)
	}

	if lastTx != nil {
		ctx, cancel := context.WithTimeout(context.Background(), TX_TIMEOUT)
		defer cancel()
		if _, err := bind.WaitMined(ctx, backend, lastTx); err != nil {
			fmt.Printf("Waiting for transactions to be mined failed: %v\n", err.Error())
		}
	}
	return nil
}
