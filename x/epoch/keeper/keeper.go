package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kiichain/kiichain3/x/epoch/types"
)

// Keeper is the epoch keeper struct
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	memKey     sdk.StoreKey
	paramstore paramtypes.Subspace
	hooks      types.ExpectedEpochHooks
}

// NewKeeper returns a new epoch keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

// SetHooks set the hooks for the epoch keeper
func (k *Keeper) SetHooks(eh types.ExpectedEpochHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set epochs hooks twice")
	}

	k.hooks = eh
	return k
}

// UnsafeSetHooks set the epoch hooks with no validation
// this is unsafe and should only be used for tests
func (k *Keeper) UnsafeSetHooks(eh types.ExpectedEpochHooks) *Keeper {
	k.hooks = eh
	return k
}

func (k *Keeper) GetParamSubspace() paramtypes.Subspace {
	return k.paramstore
}
