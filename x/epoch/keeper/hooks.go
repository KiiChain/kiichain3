package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/x/epoch/types"
)

// AfterEpochEnd is the keeper execution of the AfterEpochEnd
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epoch types.Epoch) {
	// Get max allowed gas
	maxHooksGasAllowed := k.GetParams(ctx).MaxHooksGasAllowed

	// Execute the hooks
	k.hooks.AfterEpochEnd(ctx, epoch, maxHooksGasAllowed)
}

// AfterEpochEnd is the keeper execution of the BeforeEpochStart
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epoch types.Epoch) {
	// Get max allowed gas
	maxHooksGasAllowed := k.GetParams(ctx).MaxHooksGasAllowed

	// Execute the hooks
	k.hooks.BeforeEpochStart(ctx, epoch, maxHooksGasAllowed)
}
