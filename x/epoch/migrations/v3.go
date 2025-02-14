package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain3/x/epoch/keeper"
	"github.com/kiichain/kiichain3/x/epoch/types"
)

// V3MigrateStore apply the migration from v2 to v3 for the module
func V3MigrateStore(ctx sdk.Context, k *keeper.Keeper) error {
	// Set the new parameter with its default value
	defaultParams := types.DefaultParams()
	k.SetParams(ctx, defaultParams)

	ctx.Logger().Info("Migration to v3 completed successfully")

	return nil
}
