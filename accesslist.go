package txfuzz

import (
	"context"
	"math/rand"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// CreateAccessList creates a new access list for a transaction via the eth_createAccessList.
func CreateAccessList(client *rpc.Client, tx *types.Transaction, from common.Address) (*types.AccessList, error) {
	msg := ethereum.CallMsg{
		From:       from,
		To:         tx.To(),
		Gas:        tx.Gas(),
		GasPrice:   tx.GasPrice(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: nil,
	}
	if client == nil {
		return &types.AccessList{}, nil
	}
	geth := gethclient.New(client)
	al, _, _, err := geth.CreateAccessList(context.Background(), msg)
	return al, err
}

type accessListMutator func(list *types.AccessList) *types.AccessList

var mutators = []accessListMutator{
	noChange,
	delete,
	addRandom,
	replaceRandom,
	replaceRandomSlot,
	fullyRandom,
}

// MutateAccessList mutates the given access list.
func MutateAccessList(list types.AccessList) *types.AccessList {
	index := rand.Intn(len(mutators))
	return mutators[index](&list)
}

// Leave the accesslist as is
func noChange(list *types.AccessList) *types.AccessList { return list }

// empty the access list
func delete(list *types.AccessList) *types.AccessList { return &types.AccessList{} }

// add a random entry and random slots to the list
func addRandom(list *types.AccessList) *types.AccessList {
	addr := randomAddress()
	keys := []common.Hash{}
	for i := 0; i < rand.Intn(10); i++ {
		h := randomHash()
		keys = append(keys, h)
	}
	tuple := types.AccessTuple{Address: addr, StorageKeys: keys}
	newList := types.AccessList(append([]types.AccessTuple{tuple}, *list...))
	return &newList
}

// replace a random entry and random slots of it in the list
func replaceRandom(list *types.AccessList) *types.AccessList {
	slot := (*list)[rand.Int31n(int32(len(*list)))]
	addr := randomAddress()
	keys := []common.Hash{}
	if len(slot.StorageKeys) == 0 {
		return list
	}
	for i := 0; i < rand.Intn(len(slot.StorageKeys)); i++ {
		h := randomHash()
		keys = append(keys, h)
	}
	tuple := types.AccessTuple{Address: addr, StorageKeys: keys}
	newList := types.AccessList(append([]types.AccessTuple{tuple}, *list...))
	return &newList
}

// replace a random slot in an existing entry
func replaceRandomSlot(list *types.AccessList) *types.AccessList {
	keyIdx := rand.Int31n(int32(len(*list)))
	slotIdx := rand.Int31n(int32(len((*list)[keyIdx].StorageKeys)))
	h := randomHash()
	(*list)[keyIdx].StorageKeys[slotIdx] = h
	return list
}

func fullyRandom(list *types.AccessList) *types.AccessList {
	var accesslist []types.AccessTuple
	for i := 0; i < rand.Int(); i++ {
		addr := randomAddress()
		keys := []common.Hash{}
		// create a fully random access list
		for q := 0; q < rand.Int(); q++ {
			h := randomHash()
			keys = append(keys, h)
		}
		tuple := types.AccessTuple{Address: addr, StorageKeys: keys}
		accesslist = append(accesslist, tuple)
	}
	newList := types.AccessList(accesslist)
	return &newList
}
