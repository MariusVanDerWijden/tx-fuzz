package txfuzz

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/MariusVanDerWijden/FuzzyVM/generator"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
)

// BlobSidecarVersion selects the wire format used for blob transaction
// sidecars. The zero value is the format required by current nodes.
type BlobSidecarVersion byte

const (
	// SidecarLatest attaches EIP-7594 cell proofs (sidecar version 1). This is
	// the only format accepted from Osaka onwards.
	SidecarLatest BlobSidecarVersion = iota

	// SidecarLegacy attaches a single proof per blob (sidecar version 0). Use
	// this to target pre-Osaka nodes.
	SidecarLegacy
)

// RandomCode creates a random byte code from the passed filler.
func RandomCode(f *filler.Filler) []byte {
	_, code := generator.GenerateProgram(f)
	return code
}

// TxOpts collects the parameters shared by all random transaction creators.
type TxOpts struct {
	Sender common.Address // address the transaction is sent from
	Nonce  uint64         // nonce of the sender

	GasPrice *big.Int // gas price to use, queried from the rpc if nil
	ChainID  *big.Int // chain id to use, queried from the rpc if nil
	GasLimit uint64   // gas limit to use, estimated via the rpc if zero

	AccessList bool // whether to attach a non-empty access list

	// AuthKey signs the EIP-7702 authorization tuples. If it is nil, no
	// set code transactions are generated.
	AuthKey *ecdsa.PrivateKey

	// AuthNonce is the current nonce of AuthKey's account. It is ignored when
	// AuthKey belongs to Sender, in which case the nonce is derived from Nonce.
	AuthNonce uint64

	// SidecarVersion selects the sidecar format of blob transactions.
	SidecarVersion BlobSidecarVersion
}

// RandomTx creates a random transaction.
func RandomTx(f *filler.Filler) (*types.Transaction, error) {
	return RandomTxWithOpts(nil, f, TxOpts{
		Nonce:    uint64(rand.Int63()),
		GasPrice: big.NewInt(rand.Int63()),
		ChainID:  big.NewInt(rand.Int63()),
	})
}

// RandomValidTx creates a random valid transaction.
// It does not mean that the transaction will succeed, but that it is well-formed.
// If gasPrice is not set, we will try to get it from the rpc
// If chainID is not set, we will try to get it from the rpc
func RandomValidTx(rpc *rpc.Client, f *filler.Filler, sender common.Address, nonce uint64, gasPrice, chainID *big.Int, al bool) (*types.Transaction, error) {
	return RandomTxWithOpts(rpc, f, TxOpts{
		Sender:     sender,
		Nonce:      nonce,
		GasPrice:   gasPrice,
		ChainID:    chainID,
		AccessList: al,
	})
}

// RandomTxWithOpts creates a random well-formed transaction of any type that
// shares a transaction pool with plain transactions, that is anything but a blob
// transaction. Blob transactions are kept out on purpose: a node keeps them in a
// separate pool and refuses to hold both a blob and a non-blob transaction for
// the same account, so mixing the two from one sender only produces rejections.
// Use RandomBlobTxWithOpts for those.
func RandomTxWithOpts(rpc *rpc.Client, f *filler.Filler, opts TxOpts) (*types.Transaction, error) {
	conf, err := newTxConf(rpc, f, opts)
	if err != nil {
		return nil, err
	}
	strategies := conf.applicable(txStrategies)
	return conf.build(strategies[int(f.Byte())%len(strategies)])
}

// RandomBlobTx creates a random blob transaction.
func RandomBlobTx(rpc *rpc.Client, f *filler.Filler, sender common.Address, nonce uint64, gasPrice, chainID *big.Int, al bool) (*types.Transaction, error) {
	return RandomBlobTxWithOpts(rpc, f, TxOpts{
		Sender:     sender,
		Nonce:      nonce,
		GasPrice:   gasPrice,
		ChainID:    chainID,
		AccessList: al,
	})
}

// RandomBlobTxWithOpts creates a random blob transaction.
func RandomBlobTxWithOpts(rpc *rpc.Client, f *filler.Filler, opts TxOpts) (*types.Transaction, error) {
	conf, err := newTxConf(rpc, f, opts)
	if err != nil {
		return nil, err
	}
	strategies := conf.applicable(blobStrategies)
	return conf.build(strategies[int(f.Byte())%len(strategies)])
}

// RandomSetCodeTxWithOpts creates a random EIP-7702 set code transaction with a
// random authorization list signed by opts.AuthKey, which must be set.
func RandomSetCodeTxWithOpts(rpc *rpc.Client, f *filler.Filler, opts TxOpts) (*types.Transaction, error) {
	if opts.AuthKey == nil {
		return nil, fmt.Errorf("set code transactions need an authorization key")
	}
	conf, err := newTxConf(rpc, f, opts)
	if err != nil {
		return nil, err
	}
	if opts.AccessList {
		return conf.build(setCodeStrategy(true))
	}
	return conf.build(setCodeStrategy(false))
}

// RandomAuthTx creates a random EIP-7702 set code transaction with the given
// authorization list.
func RandomAuthTx(rpc *rpc.Client, f *filler.Filler, sender common.Address, nonce uint64, gasPrice, chainID *big.Int, al bool, aList []types.SetCodeAuthorization) (*types.Transaction, error) {
	conf, err := newTxConf(rpc, f, TxOpts{
		Sender:     sender,
		Nonce:      nonce,
		GasPrice:   gasPrice,
		ChainID:    chainID,
		AccessList: al,
	})
	if err != nil {
		return nil, err
	}
	conf.authList = aList
	return conf.build(setCodeStrategy(al))
}

// setCodeStrategy returns the set code strategy with or without an access list.
func setCodeStrategy(accessList bool) txCreationStrategy {
	for _, strategy := range txStrategies {
		if strategy.needsAuthList && strategy.needsAccessList == accessList {
			return strategy
		}
	}
	panic("no set code strategy registered")
}

type txConf struct {
	rpc      *rpc.Client
	f        *filler.Filler
	nonce    uint64
	sender   common.Address
	to       *common.Address
	value    *big.Int
	gasLimit uint64
	gasPrice *big.Int
	chainID  *big.Int
	code     []byte

	accessList     bool
	authList       []types.SetCodeAuthorization
	sidecarVersion BlobSidecarVersion
}

func newTxConf(client *rpc.Client, f *filler.Filler, opts TxOpts) (*txConf, error) {
	var (
		gasPrice = opts.GasPrice
		chainID  = opts.ChainID
		gasLimit = opts.GasLimit
		to       = fillerAddress(f)
		value    = big.NewInt(0)
		code     = RandomCode(f)
	)
	if len(code) > 128 {
		code = code[:128]
	}
	if gasLimit == 0 {
		gasLimit = 100_000
	}
	if client != nil {
		backend := ethclient.NewClient(client)
		var err error
		if gasPrice == nil {
			if gasPrice, err = backend.SuggestGasPrice(context.Background()); err != nil {
				log.Warn("Error suggesting gas price", "err", err)
				gasPrice = big.NewInt(1)
			}
		}
		if chainID == nil {
			if chainID, err = backend.ChainID(context.Background()); err != nil {
				log.Warn("Error fetching chain id", "err", err)
				chainID = big.NewInt(1)
			}
		}
		// Try to estimate the gas, fall back to the default on failure. The
		// estimate is capped at the EIP-7825 per-transaction limit, above which
		// a transaction can never be included.
		if opts.GasLimit == 0 {
			// Naming both a gas price and the 1559 fee caps is an error, the
			// estimate never came back while this passed all three.
			gas, err := backend.EstimateGas(context.Background(), ethereum.CallMsg{
				From:     opts.Sender,
				To:       &to,
				Gas:      params.MaxTxGas,
				GasPrice: gasPrice,
				Value:    value,
				Data:     code,
			})
			if err != nil {
				log.Warn("Error estimating gas", "err", err)
			} else {
				gasLimit = gas
			}
		}
	}
	if gasPrice == nil {
		return nil, fmt.Errorf("no gas price available")
	}
	if chainID == nil {
		return nil, fmt.Errorf("no chain id available")
	}
	conf := &txConf{
		rpc:            client,
		f:              f,
		nonce:          opts.Nonce,
		sender:         opts.Sender,
		to:             &to,
		value:          value,
		gasLimit:       min(gasLimit, params.MaxTxGas),
		gasPrice:       gasPrice,
		chainID:        chainID,
		code:           code,
		accessList:     opts.AccessList,
		sidecarVersion: opts.SidecarVersion,
	}
	// Set code transactions are only generated if we can sign authorizations.
	if opts.AuthKey != nil {
		authNonce := opts.AuthNonce
		if crypto.PubkeyToAddress(opts.AuthKey.PublicKey) == opts.Sender {
			// The authority also sends the transaction, so its nonce has
			// already been bumped by the transaction itself by the time the
			// authorization is applied.
			authNonce = opts.Nonce + 1
		}
		authList, err := RandomAuthList(f, opts.AuthKey, uint256.MustFromBig(chainID), authNonce)
		if err != nil {
			return nil, err
		}
		conf.authList = authList
	}
	return conf, nil
}

// txCreationStrategy is one way of building a random transaction.
type txCreationStrategy struct {
	name string
	// needsAuthList marks a strategy that builds a set code transaction, which
	// is only well-formed with a non-empty authorization list.
	needsAuthList bool
	// needsAccessList marks a strategy that asks the node to build an access
	// list, which requires an rpc connection.
	needsAccessList bool
	create          func(conf *txConf) (*types.Transaction, error)
}

// txStrategies lists every way of building a transaction that shares the plain
// transaction pool. Blob transactions live in blobStrategies.
var txStrategies = []txCreationStrategy{
	{name: "legacyContractCreation", create: legacyContractCreation},
	{name: "legacyTx", create: legacyTx},
	{name: "emptyAlContractCreation", create: emptyAlContractCreation},
	{name: "emptyAlTx", create: emptyAlTx},
	{name: "contractCreation1559", create: contractCreation1559},
	{name: "tx1559", create: tx1559},
	{name: "emptyAlSetCodeTx", needsAuthList: true, create: emptyAlSetCodeTx},
	{name: "fullAl1559ContractCreation", needsAccessList: true, create: fullAl1559ContractCreation},
	{name: "fullAl1559Tx", needsAccessList: true, create: fullAl1559Tx},
	{name: "fullAlContractCreation", needsAccessList: true, create: fullAlContractCreation},
	{name: "fullAlTx", needsAccessList: true, create: fullAlTx},
	{name: "fullAlSetCodeTx", needsAuthList: true, needsAccessList: true, create: fullAlSetCodeTx},
}

// blobStrategies lists the ways of building a blob transaction.
var blobStrategies = []txCreationStrategy{
	{name: "emptyAlBlobTx", create: emptyAlBlobTx},
	{name: "fullAlBlobTx", needsAccessList: true, create: fullAlBlobTx},
}

// build creates a transaction with the given strategy, and makes sure its gas
// limit covers the intrinsic cost.
//
// The intrinsic cost cannot be known before the transaction exists: it depends
// on whether the strategy deploys a contract, how many authorizations it
// carries and how large an access list the node handed back. Strategies read
// the configuration without touching the filler, so the transaction can simply
// be built a second time with a corrected gas limit.
func (conf *txConf) build(strategy txCreationStrategy) (*types.Transaction, error) {
	tx, err := strategy.create(conf)
	if err != nil {
		return nil, err
	}
	floor, err := intrinsicGas(conf.sender, tx)
	if err != nil {
		return nil, err
	}
	if tx.Gas() >= floor {
		return tx, nil
	}
	// Leave the originally requested gas limit as headroom for execution on top
	// of the intrinsic cost, so the transaction does more than just pay for
	// itself.
	gasLimit := min(floor+conf.gasLimit, params.MaxTxGas)
	if gasLimit < floor {
		return nil, fmt.Errorf("intrinsic gas %d exceeds the per-transaction gas cap %d", floor, params.MaxTxGas)
	}
	conf.gasLimit = gasLimit
	return strategy.create(conf)
}

// intrinsicRules are the rules of the newest fork the vendored go-ethereum
// activates in a dev chain. The intrinsic gas requirement of a transaction only
// grows with forks, so sizing gas limits with these is safe on any older chain.
var intrinsicRules = params.AllDevChainProtocolChanges.Rules(common.Big0, true, ^uint64(0))

// intrinsicGas returns the smallest gas limit that covers the transaction's
// intrinsic cost. A transaction below it is rejected with "intrinsic gas too
// low" before it ever executes.
func intrinsicGas(sender common.Address, tx *types.Transaction) (uint64, error) {
	value, overflow := uint256.FromBig(tx.Value())
	if overflow {
		return 0, fmt.Errorf("transaction value overflows 256 bits")
	}
	gas, err := core.IntrinsicGas(tx.Data(), tx.AccessList(), tx.SetCodeAuthorizations(), sender, tx.To(), value, intrinsicRules)
	if err != nil {
		return 0, err
	}
	// Since EIP-7623 a transaction also has to cover the calldata floor.
	floor, err := core.FloorDataGas(intrinsicRules, sender, tx.To(), value, tx.Data(), tx.AccessList())
	if err != nil {
		return 0, err
	}
	return max(gas, floor), nil
}

// applicable filters the strategies down to the ones this configuration can
// actually build a transaction with.
func (conf *txConf) applicable(strategies []txCreationStrategy) []txCreationStrategy {
	filtered := make([]txCreationStrategy, 0, len(strategies))
	for _, strategy := range strategies {
		if strategy.needsAccessList && !conf.accessList {
			continue
		}
		if strategy.needsAuthList && len(conf.authList) == 0 {
			continue
		}
		filtered = append(filtered, strategy)
	}
	return filtered
}

func legacyContractCreation(conf *txConf) (*types.Transaction, error) {
	// Legacy contract creation
	return types.NewContractCreation(conf.nonce, conf.value, conf.gasLimit, conf.gasPrice, conf.code), nil
}

func legacyTx(conf *txConf) (*types.Transaction, error) {
	// Legacy transaction
	return types.NewTransaction(conf.nonce, *conf.to, conf.value, conf.gasLimit, conf.gasPrice, conf.code), nil
}

func emptyAlContractCreation(conf *txConf) (*types.Transaction, error) {
	// AccessList contract creation
	return newALTx(conf.nonce, nil, conf.gasLimit, conf.chainID, conf.gasPrice, conf.value, conf.code, make(types.AccessList, 0)), nil
}

func emptyAlTx(conf *txConf) (*types.Transaction, error) {
	// AccessList transaction
	return newALTx(conf.nonce, conf.to, conf.gasLimit, conf.chainID, conf.gasPrice, conf.value, conf.code, make(types.AccessList, 0)), nil
}

func contractCreation1559(conf *txConf) (*types.Transaction, error) {
	// 1559 contract creation
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	return new1559Tx(conf.nonce, nil, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, make(types.AccessList, 0)), nil
}

func tx1559(conf *txConf) (*types.Transaction, error) {
	// 1559 transaction
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	return new1559Tx(conf.nonce, conf.to, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, make(types.AccessList, 0)), nil
}

func fullAlContractCreation(conf *txConf) (*types.Transaction, error) {
	// AccessList contract creation with AL
	tx := types.NewContractCreation(conf.nonce, conf.value, conf.gasLimit, conf.gasPrice, conf.code)
	al, err := CreateAccessList(conf.rpc, tx, conf.sender)
	if err != nil {
		return nil, err
	}
	return newALTx(conf.nonce, nil, conf.gasLimit, conf.chainID, conf.gasPrice, conf.value, conf.code, *al), nil
}

func fullAlTx(conf *txConf) (*types.Transaction, error) {
	// AccessList transaction with AL
	tx := types.NewTransaction(conf.nonce, *conf.to, conf.value, conf.gasLimit, conf.gasPrice, conf.code)
	al, err := CreateAccessList(conf.rpc, tx, conf.sender)
	if err != nil {
		return nil, err
	}
	return newALTx(conf.nonce, conf.to, conf.gasLimit, conf.chainID, conf.gasPrice, conf.value, conf.code, *al), nil
}

func fullAl1559ContractCreation(conf *txConf) (*types.Transaction, error) {
	// 1559 contract creation with AL
	tx := types.NewContractCreation(conf.nonce, conf.value, conf.gasLimit, conf.gasPrice, conf.code)
	al, err := CreateAccessList(conf.rpc, tx, conf.sender)
	if err != nil {
		return nil, err
	}
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	return new1559Tx(conf.nonce, nil, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, *al), nil
}

func fullAl1559Tx(conf *txConf) (*types.Transaction, error) {
	// 1559 tx with AL
	tx := types.NewTransaction(conf.nonce, *conf.to, conf.value, conf.gasLimit, conf.gasPrice, conf.code)
	al, err := CreateAccessList(conf.rpc, tx, conf.sender)
	if err != nil {
		return nil, err
	}
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	return new1559Tx(conf.nonce, conf.to, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, *al), nil
}

func emptyAlSetCodeTx(conf *txConf) (*types.Transaction, error) {
	// 7702 transaction without AL
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	return New7702Tx(conf.nonce, *conf.to, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, make(types.AccessList, 0), conf.authList), nil
}

func fullAlSetCodeTx(conf *txConf) (*types.Transaction, error) {
	// 7702 transaction with AL
	tx := types.NewTransaction(conf.nonce, *conf.to, conf.value, conf.gasLimit, conf.gasPrice, conf.code)
	al, err := CreateAccessList(conf.rpc, tx, conf.sender)
	if err != nil {
		return nil, err
	}
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	return New7702Tx(conf.nonce, *conf.to, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, *al, conf.authList), nil
}

func emptyAlBlobTx(conf *txConf) (*types.Transaction, error) {
	// 4844 transaction without AL
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	data, err := randomBlobData(conf.f)
	if err != nil {
		return nil, err
	}
	blobFeeCap, err := getBlobCap(conf.rpc)
	if err != nil {
		return nil, err
	}
	return New4844Tx(conf.nonce, conf.to, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, blobFeeCap, data, make(types.AccessList, 0), conf.sidecarVersion)
}

func fullAlBlobTx(conf *txConf) (*types.Transaction, error) {
	// 4844 transaction with AL
	tx := types.NewTransaction(conf.nonce, *conf.to, conf.value, conf.gasLimit, conf.gasPrice, conf.code)
	al, err := CreateAccessList(conf.rpc, tx, conf.sender)
	if err != nil {
		return nil, err
	}
	tip, feecap, err := getCaps(conf.rpc, conf.gasPrice)
	if err != nil {
		return nil, err
	}
	data, err := randomBlobData(conf.f)
	if err != nil {
		return nil, err
	}
	blobFeeCap, err := getBlobCap(conf.rpc)
	if err != nil {
		return nil, err
	}
	return New4844Tx(conf.nonce, conf.to, conf.gasLimit, conf.chainID, tip, feecap, conf.value, conf.code, blobFeeCap, data, *al, conf.sidecarVersion)
}

func newALTx(nonce uint64, to *common.Address, gasLimit uint64, chainID, gasPrice, value *big.Int, code []byte, al types.AccessList) *types.Transaction {
	return types.NewTx(&types.AccessListTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasPrice:   gasPrice,
		Gas:        gasLimit,
		To:         to,
		Value:      value,
		Data:       code,
		AccessList: al,
	})
}

func new1559Tx(nonce uint64, to *common.Address, gasLimit uint64, chainID, tip, feeCap, value *big.Int, code []byte, al types.AccessList) *types.Transaction {
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  tip,
		GasFeeCap:  feeCap,
		Gas:        gasLimit,
		To:         to,
		Value:      value,
		Data:       code,
		AccessList: al,
	})
}

// New4844Tx creates a blob transaction carrying blobData, encoded into as many
// blobs as it needs.
func New4844Tx(nonce uint64, to *common.Address, gasLimit uint64, chainID, tip, feeCap, value *big.Int, code []byte, blobFeeCap *big.Int, blobData []byte, al types.AccessList, version BlobSidecarVersion) (*types.Transaction, error) {
	sidecar, err := EncodeBlobs(blobData, version)
	if err != nil {
		return nil, err
	}
	return types.NewTx(&types.BlobTx{
		ChainID:    uint256.MustFromBig(chainID),
		Nonce:      nonce,
		GasTipCap:  uint256.MustFromBig(tip),
		GasFeeCap:  uint256.MustFromBig(feeCap),
		Gas:        gasLimit,
		To:         *to,
		Value:      uint256.MustFromBig(value),
		Data:       code,
		AccessList: al,
		BlobFeeCap: uint256.MustFromBig(blobFeeCap),
		BlobHashes: sidecar.BlobHashes(),
		Sidecar:    sidecar,
	}), nil
}

func New7702Tx(nonce uint64, to common.Address, gasLimit uint64, chainID, tip, feeCap, value *big.Int, code []byte, al types.AccessList, auth []types.SetCodeAuthorization) *types.Transaction {
	return types.NewTx(
		&types.SetCodeTx{
			ChainID:    uint256.MustFromBig(chainID),
			Nonce:      nonce,
			To:         to,
			GasTipCap:  uint256.MustFromBig(tip),
			GasFeeCap:  uint256.MustFromBig(feeCap),
			Gas:        gasLimit,
			Value:      uint256.MustFromBig(value),
			Data:       code,
			AuthList:   auth,
			AccessList: al,
		},
	)
}

func getCaps(rpc *rpc.Client, defaultGasPrice *big.Int) (*big.Int, *big.Int, error) {
	if rpc == nil {
		// Without a node to ask, spend the whole gas price on the fee cap and
		// take the tip out of it. Subtracting the tip from the fee cap instead
		// would push the tip above the cap for any gas price below two gwei,
		// which no node accepts.
		tip := big.NewInt(params.GWei)
		if tip.Cmp(defaultGasPrice) > 0 {
			tip = new(big.Int).Set(defaultGasPrice)
		}
		return tip, defaultGasPrice, nil
	}
	client := ethclient.NewClient(rpc)
	tip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, nil, err
	}
	feeCap, err := client.SuggestGasPrice(context.Background())
	return tip, feeCap, err
}

// blobFeeCapMultiplier is the factor the current blob base fee is multiplied
// with, to leave room for the blob base fee to rise while the transaction sits
// in the pool.
const blobFeeCapMultiplier = 4

// getBlobCap returns the blob fee cap to use for a blob transaction.
func getBlobCap(client *rpc.Client) (*big.Int, error) {
	if client == nil {
		return big.NewInt(params.BlobTxMinBlobGasprice), nil
	}
	blobBaseFee, err := ethclient.NewClient(client).BlobBaseFee(context.Background())
	if err != nil {
		// Not every node exposes eth_blobBaseFee, fall back to a value that
		// covers the minimum blob gas price.
		log.Warn("Error fetching blob base fee", "err", err)
		return big.NewInt(1_000_000), nil
	}
	feeCap := new(big.Int).Mul(blobBaseFee, big.NewInt(blobFeeCapMultiplier))
	if feeCap.Cmp(big.NewInt(params.BlobTxMinBlobGasprice)) < 0 {
		feeCap = big.NewInt(params.BlobTxMinBlobGasprice)
	}
	return feeCap, nil
}

func encodeBlobs(data []byte) []kzg4844.Blob {
	blobs := []kzg4844.Blob{{}}
	blobIndex := 0
	fieldIndex := -1
	for i := 0; i < len(data); i += 31 {
		fieldIndex++
		if fieldIndex == params.BlobTxFieldElementsPerBlob {
			blobs = append(blobs, kzg4844.Blob{})
			blobIndex++
			fieldIndex = 0
		}
		max := i + 31
		if max > len(data) {
			max = len(data)
		}
		copy(blobs[blobIndex][fieldIndex*32+1:], data[i:max])
	}
	return blobs
}

// EncodeBlobs packs data into blobs and returns the matching sidecar in the
// requested version.
func EncodeBlobs(data []byte, version BlobSidecarVersion) (*types.BlobTxSidecar, error) {
	blobs := encodeBlobs(data)
	if len(blobs) > params.BlobTxMaxBlobs {
		return nil, fmt.Errorf("too much blob data: %d bytes need %d blobs, limit is %d", len(data), len(blobs), params.BlobTxMaxBlobs)
	}
	commits := make([]kzg4844.Commitment, 0, len(blobs))
	for i := range blobs {
		commit, err := kzg4844.BlobToCommitment(&blobs[i])
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}
	var proofs []kzg4844.Proof
	switch version {
	case SidecarLegacy:
		// One proof covering the whole blob (pre-Osaka).
		for i := range blobs {
			proof, err := kzg4844.ComputeBlobProof(&blobs[i], commits[i])
			if err != nil {
				return nil, err
			}
			proofs = append(proofs, proof)
		}
		return types.NewBlobTxSidecar(types.BlobSidecarVersion0, blobs, commits, proofs), nil
	case SidecarLatest:
		// EIP-7594: CellProofsPerBlob proofs per blob, so that the blob can be
		// reconstructed from a subset of its cells.
		for i := range blobs {
			cellProofs, err := kzg4844.ComputeCellProofs(&blobs[i])
			if err != nil {
				return nil, err
			}
			proofs = append(proofs, cellProofs...)
		}
		return types.NewBlobTxSidecar(types.BlobSidecarVersion1, blobs, commits, proofs), nil
	default:
		return nil, fmt.Errorf("unknown blob sidecar version %d", version)
	}
}

// SignTx signs the transaction with the newest signer, which covers every
// transaction type this package can generate.
func SignTx(tx *types.Transaction, chainID *big.Int, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	return types.SignTx(tx, types.NewPragueSigner(chainID), key)
}
