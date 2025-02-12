package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/mint/keeper"
	"github.com/kiichain/kiichain3/x/mint/types"
)

// HandleUpdateMinterProposal handle the update minter governance proposal
func HandleUpdateMinterProposal(ctx sdk.Context, k *keeper.Keeper, p *types.UpdateMinterProposal) error {
	// Validate the minter object
	// This validates mint total, mint dates and mint denom
	err := types.ValidateMinter(*p.Minter)
	if err != nil {
		return err
	}

	// Validate the inflation rate
	if err := k.ValidateInflationRate(ctx, p.Minter); err != nil {
		return err
	}

	// Set the minter
	k.SetMinter(ctx, *p.Minter)
	return nil
}
