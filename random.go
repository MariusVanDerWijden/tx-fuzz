package txfuzz

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	mathRand "math/rand"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

const (
	// blobBytes is the number of bytes that can be packed into a single blob.
	// Only 31 of the 32 bytes of a field element are usable, the first byte is
	// zeroed to keep the element below the BLS modulus.
	blobBytes = params.BlobTxFieldElementsPerBlob * 31

	// maxDataPerTx is the maximum amount of blob data attached to a single
	// transaction, bounded by the per-transaction blob limit.
	maxDataPerTx = params.BlobTxMaxBlobs * blobBytes

	// maxAuthorizations bounds the number of authorizations in a generated
	// EIP-7702 authorization list. The protocol imposes no limit, but an
	// unbounded list makes the transaction too large to ever be accepted.
	maxAuthorizations = 16
)

// specialAddresses are addresses with protocol-level meaning. Hitting one of
// them is far more interesting than hitting a random account, so they are
// sampled with a much higher probability than uniform.
var specialAddresses = []common.Address{
	{},
	common.HexToAddress(ADDR),
	params.BeaconRootsAddress,
	params.WithdrawalQueueAddress,
	params.ConsolidationQueueAddress,
	params.SystemAddress,
	params.HistoryStorageAddress,
	params.DeterministicFactoryAddress,
}

func randomHash() common.Hash {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return common.BytesToHash(b)
}

func randomAddress() common.Address {
	// With probability 3/8 return a fully random address, otherwise one of the
	// special addresses.
	if mathRand.Int31n(8) < 3 {
		b := make([]byte, 20)
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		return common.BytesToAddress(b)
	}
	return specialAddresses[mathRand.Intn(len(specialAddresses))]
}

// fillerAddress returns an address derived from the filler, so that it is
// reproducible from the seed.
func fillerAddress(f *filler.Filler) common.Address {
	if f.Byte()%8 < 3 {
		return common.BytesToAddress(f.ByteSlice(20))
	}
	return specialAddresses[int(f.Byte())%len(specialAddresses)]
}

// randomBlobData returns blob data filling 1 up to BlobTxMaxBlobs blobs. The
// contents are derived from the filler, so a run is reproducible from its seed.
func randomBlobData(f *filler.Filler) ([]byte, error) {
	blobs := 1 + int(f.Byte())%params.BlobTxMaxBlobs
	size := blobs * blobBytes
	// Fill the blobs completely most of the time; a partially filled trailing
	// blob exercises the zero padding path.
	if f.Byte() < 64 {
		size -= 1 + int(f.Uint32())%blobBytes
	}
	if size <= 0 || size > maxDataPerTx {
		return nil, fmt.Errorf("invalid blob data size %d", size)
	}
	return f.ByteSlice(size), nil
}

// randomAuthEntry creates an unsigned authorization tuple.
//
// A chain ID of 0 makes the authorization valid on every chain and is, besides
// the actual chain ID, the only value a node will accept, so both are sampled
// deliberately. Likewise, an authorization only applies if its nonce matches the
// authority's current nonce, so the caller passes that in rather than leaving it
// to chance.
func randomAuthEntry(f *filler.Filler, chainID *uint256.Int, authority common.Address, nonce uint64) types.SetCodeAuthorization {
	auth := types.SetCodeAuthorization{
		ChainID: *chainID,
		Address: fillerAddress(f),
		Nonce:   nonce,
	}
	switch f.Byte() % 8 {
	case 0:
		// Universal authorization, valid on any chain.
		auth.ChainID = *uint256.NewInt(0)
	case 1:
		// Wrong chain, must be ignored by the node.
		auth.ChainID = *uint256.NewInt(f.Uint64())
	case 2:
		// Wrong nonce, must be ignored by the node.
		auth.Nonce = f.Uint64()
	case 3:
		// Delegate the authority to itself.
		auth.Address = authority
	case 4:
		// Clear an existing delegation.
		auth.Address = common.Address{}
	}
	return auth
}

// RandomAuthList creates a random, signed EIP-7702 authorization list. The
// tuples are signed by sk, so nonce must be the current nonce of sk's account
// for them to take effect.
//
// Applying an authorization bumps the authority's nonce, so every tuple that is
// meant to apply is given the next nonce in sequence. Only tuples the generator
// deliberately corrupts deviate from that.
func RandomAuthList(f *filler.Filler, sk *ecdsa.PrivateKey, chainID *uint256.Int, nonce uint64) ([]types.SetCodeAuthorization, error) {
	var (
		entries   = 1 + int(f.Byte())%maxAuthorizations
		authority = crypto.PubkeyToAddress(sk.PublicKey)
		authList  = make([]types.SetCodeAuthorization, 0, entries)
	)
	for i := 0; i < entries; i++ {
		signed, err := types.SignSetCode(sk, randomAuthEntry(f, chainID, authority, nonce+uint64(i)))
		if err != nil {
			return nil, err
		}
		authList = append(authList, signed)
	}
	return authList, nil
}
