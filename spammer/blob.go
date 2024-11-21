package spammer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/MariusVanDerWijden/tx-fuzz/helper"
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
		tx, err := txfuzz.RandomBlobTx(config.backend, f, sender, nonce, nil, nil, config.accessList)
		if err != nil {
			log.Warn("Could not create valid tx: %v", nonce)
			return err
		}

		signedTx, err := types.SignTx(tx, types.NewCancunSigner(chainID), key)
		if err != nil {
			log.Warn("Could not sign tx: %v", err)
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

func SendBasicBlobTransactions(config *Config, key *ecdsa.PrivateKey, f *filler.Filler) error {
	backend := ethclient.NewClient(config.backend)
	sender := crypto.PubkeyToAddress(key.PublicKey)
	chainID, err := backend.ChainID(context.Background())
	if err != nil {
		log.Warn("Could not get chainID, using default")
		chainID = big.NewInt(0x01000666)
	}

	// Deploy blob caller
	bytecode := "608060405234801561001057600080fd5b5061047b806100206000396000f3fe608060405234801561001057600080fd5b5060003660606000600a90506000808273ffffffffffffffffffffffffffffffffffffffff1661c350878760405161004992919061010a565b60006040518083038160008787f1925050503d8060008114610087576040519150601f19603f3d011682016040523d82523d6000602084013e61008c565b606091505b5091509150809350816000806101000a81548160ff02191690831515021790555080600190816100bc9190610373565b50505050915050805190602001f35b600081905092915050565b82818337600083830152505050565b60006100f183856100cb565b93506100fe8385846100d6565b82840190509392505050565b60006101178284866100e5565b91508190509392505050565b600081519050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806101a457607f821691505b6020821081036101b7576101b661015d565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b60006008830261021f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826101e2565b61022986836101e2565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b600061027061026b61026684610241565b61024b565b610241565b9050919050565b6000819050919050565b61028a83610255565b61029e61029682610277565b8484546101ef565b825550505050565b600090565b6102b36102a6565b6102be818484610281565b505050565b5b818110156102e2576102d76000826102ab565b6001810190506102c4565b5050565b601f821115610327576102f8816101bd565b610301846101d2565b81016020851015610310578190505b61032461031c856101d2565b8301826102c3565b50505b505050565b600082821c905092915050565b600061034a6000198460080261032c565b1980831691505092915050565b60006103638383610339565b9150826002028217905092915050565b61037c82610123565b67ffffffffffffffff8111156103955761039461012e565b5b61039f825461018c565b6103aa8282856102e6565b600060209050601f8311600181146103dd57600084156103cb578287015190505b6103d58582610357565b86555061043d565b601f1984166103eb866101bd565b60005b82811015610413578489015182556001820191506020850194506020810190506103ee565b86831015610430578489015161042c601f891682610339565b8355505b6001600288020188555050505b50505050505056fea264697066735822122089d7332a134ee7e7d76876ef5f4e74d939f9d9d9f3344e6afb518c96fff0b63164736f6c63430008120033"
	addr, err := helper.Deploy(bytecode)
	if err != nil {
		return err
	}

	var lastTx *types.Transaction
	for i := uint64(0); i < config.N; i++ {
		nonce, err := backend.NonceAt(context.Background(), sender, big.NewInt(-1))
		if err != nil {
			return err
		}

		code := common.FromHex("01d18459b334ffe8e2226eef1db874fda6db2bdd9357268b39220af2d59464fb564c0a11a0f704f4fc3e8acfe0f8245f0ad1347b378fbf96e206da11a5d3630624d25032e67a7e6a4910df5834b8fe70e6bcfeeac0352434196bdf4b2485d5a1978a0d595c823c05947b1156175e72634a377808384256e9921ebf72181890be2d6b58d4a73a880541d1656875654806942307f266e636553e94006d11423f2688945ff3bdf515859eba1005c1a7708d620a94d91a1c0c285f9584e75ec2f82a")
		tx, err := txfuzz.RandomBlobTxWithCode(config.backend, f, sender, nonce, nil, nil, config.accessList, code, &addr)
		if err != nil {
			log.Warn("Could not create valid tx: %v", nonce)
			return err
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
