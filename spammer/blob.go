package spammer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
		var tx *types.Transaction
		if !config.Basic {
			var err error
			tx, err = txfuzz.RandomBlobTx(config.backend, f, sender, nonce, nil, nil, config.accessList)
			if err != nil {
				log.Warn("Could not create valid tx: %v", nonce)
				return err
			}
		} else {
			var err error
			code := common.FromHex("01d18459b334ffe8e2226eef1db874fda6db2bdd9357268b39220af2d59464fb564c0a11a0f704f4fc3e8acfe0f8245f0ad1347b378fbf96e206da11a5d3630624d25032e67a7e6a4910df5834b8fe70e6bcfeeac0352434196bdf4b2485d5a1978a0d595c823c05947b1156175e72634a377808384256e9921ebf72181890be2d6b58d4a73a880541d1656875654806942307f266e636553e94006d11423f2688945ff3bdf515859eba1005c1a7708d620a94d91a1c0c285f9584e75ec2f82a")
			tx, err = txfuzz.RandomBlobTxWithCode(config.backend, f, sender, nonce, nil, nil, config.accessList, code)
			if err != nil {
				log.Warn("Could not create valid tx: %v", nonce)
				return err
			}
		}

		signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainID), key)
		if err != nil {
			return err
		}
		if err := backend.SendTransaction(context.Background(), signedTx); err != nil {
			log.Warn("Could not submit transaction: %v", err)
			return err
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
