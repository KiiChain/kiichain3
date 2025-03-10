package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/x/mint/types"
)

// ValidateMinter validates a new minter inflation rate
func (k Keeper) ValidateInflationRate(ctx sdk.Context, newMinter *types.Minter) error {
	// Get the inflation rate from params
	inflationMax := k.GetParams(ctx).InflationMax

	// Get the inflation rate for the new minter
	newInflationRate, err := k.GetAnnualInflationForMint(ctx, newMinter)
	if err != nil {
		return err
	}

	// Ensure the new inflation rate does not exceed the max allowed inflation
	if newInflationRate.GT(inflationMax) {
		return fmt.Errorf("annual inflation rate %.6f exceeds maximum allowed inflation rate of %.6f", newInflationRate, inflationMax)
	}

	// Return nil if we reach here
	return nil
}

// GetAnnualInflationForMint returns the annual inflation for
func (k Keeper) GetAnnualInflationForMint(ctx sdk.Context, minter *types.Minter) (sdk.Dec, error) {
	// Get the current supply for the token that will be minted
	supply := k.bankKeeper.GetSupply(ctx, minter.Denom).Amount.ToDec()

	// Check if we have supply
	if supply.IsZero() {
		return sdk.ZeroDec(), fmt.Errorf("can't calculate inflation when supply is zero")
	}

	// Get the end date and start date
	startDate, err := minter.GetStartDateTime()
	if err != nil {
		return sdk.Dec{}, err
	}
	endDate, err := minter.GetEndDateTime()
	if err != nil {
		return sdk.Dec{}, err
	}

	// Compute total days in the minting period
	totalDays := types.DaysBetween(startDate, endDate)
	if totalDays == 0 {
		return sdk.Dec{}, fmt.Errorf("invalid minting duration: must be greater than zero days")
	}

	// Convert days to years
	years := sdk.NewDec(int64(totalDays)).Quo(sdk.NewDec(365))

	// Get the total mintable amount
	totalMint := sdk.NewDecFromInt(sdk.NewIntFromUint64(minter.TotalMintAmount))

	// Compute the annualized inflation rate
	annualInflation := totalMint.Quo(years).Quo(supply)

	// Return the annual inflation
	return annualInflation, nil
}
