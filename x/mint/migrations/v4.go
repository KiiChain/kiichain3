package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/x/mint/keeper"
	"github.com/kiichain/kiichain/x/mint/types"
)

// V4MigrateStore apply the migration from v3 to v4 for the module
func V4MigrateStore(ctx sdk.Context, k *keeper.Keeper) error {
	// Set the new parameter with its default value
	defaultParams := types.DefaultParams()
	k.SetParams(ctx, defaultParams)

	ctx.Logger().Info("Migration to v4 completed successfully")

	return nil
}
