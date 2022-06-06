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

func CreateAccessList(client *rpc.Client, tx *types.Transaction, from common.Address) (*types.AccessList, error) {
	msg := ethereum.CallMsg{
		From:       from,
		To:         tx.To(),
		Gas:        tx.Gas(),
		GasPrice:   tx.GasPrice(),
		GasFeeCap:  tx.GasFeeCap(),
		GasTipCap:  tx.GasTipCap(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: nil,
	}
	return createAccessList(client, msg)
}

func createAccessList(client *rpc.Client, msg ethereum.CallMsg) (*types.AccessList, error) {
	if client == nil {
		return &types.AccessList{}, nil
	}
	geth := gethclient.New(client)
	al, _, _, err := geth.CreateAccessList(context.Background(), msg)
	return al, err
}

func MutateAccessList(list types.AccessList) *types.AccessList {
	switch rand.Int31n(5) {
	case 0:
		// Leave the accesslist as is
		return &list
	case 1:
		// delete the access list
		return &types.AccessList{}
	case 2:
		// empty the access list
		return &types.AccessList{}
	case 3:
		// add a random entry and random slots to the list
		addr := randomAddress()
		keys := []common.Hash{}
		for i := 0; i < rand.Intn(10); i++ {
			h := randomHash()
			keys = append(keys, h)
		}
		tuple := types.AccessTuple{Address: addr, StorageKeys: keys}
		newList := types.AccessList(append([]types.AccessTuple{tuple}, list...))
		return &newList
	case 4:
		// replace a random entry and random slots of it in the list
		slot := list[rand.Int31n(int32(len(list)))]
		addr := randomAddress()
		keys := []common.Hash{}
		if len(slot.StorageKeys) == 0 {
			break
		}
		for i := 0; i < rand.Intn(len(slot.StorageKeys)); i++ {
			h := randomHash()
			keys = append(keys, h)
		}
		tuple := types.AccessTuple{Address: addr, StorageKeys: keys}
		newList := types.AccessList(append([]types.AccessTuple{tuple}, list...))
		return &newList
	case 5:
		// replace a random slot in an existing entry
		keyIdx := rand.Int31n(int32(len(list)))
		slotIdx := rand.Int31n(int32(len(list[keyIdx].StorageKeys)))
		h := randomHash()
		list[keyIdx].StorageKeys[slotIdx] = h
	case 6:
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
	return &list
}
