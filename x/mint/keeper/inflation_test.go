package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/x/mint/types"
)

// TestGetAnnualInflationForMint tests the calculation of the annual inflation rate
func TestGetAnnualInflationForMint(t *testing.T) {
	// Initialize a new test app
	app, ctx := createTestApp(false)

	// Mint 365 tokens to easy out the calculation
	err := app.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(
		sdk.NewCoin("ukii", sdk.NewInt(365_000_000)), // 365 tokens
	))
	require.NoError(t, err)

	// Prepare the test cases
	testCases := []struct {
		name        string
		minter      types.Minter
		errContains string
		expectedMin sdk.Dec // Minimum and maximum to checkout around error rates
		expectedMax sdk.Dec // Maximum expected inflation rate
	}{
		{
			name: "Good - Valid annual inflation calculation (10%)",
			minter: types.Minter{
				StartDate:       "2022-01-01", // A year of difference
				EndDate:         "2023-01-01",
				Denom:           "ukii",
				TotalMintAmount: 36_500_000, // Mint 36.5 tokens in a year
			},
			expectedMin: sdk.MustNewDecFromStr("0.099"), // ~10% inflation
			expectedMax: sdk.MustNewDecFromStr("0.101"),
		},
		{
			name: "Good - Valid annual inflation calculation (20%)",
			minter: types.Minter{
				StartDate:       "2022-01-01", // A year of difference
				EndDate:         "2023-01-01",
				Denom:           "ukii",
				TotalMintAmount: 73_000_000, // Mint 73 tokens in a year
			},
			expectedMin: sdk.MustNewDecFromStr("0.199"), // ~20% inflation
			expectedMax: sdk.MustNewDecFromStr("0.201"),
		},
		{
			name: "Good - Valid annual inflation - 10 days",
			minter: types.Minter{
				StartDate:       "2022-01-10", // 10 days of difference
				EndDate:         "2022-01-20",
				Denom:           "ukii",
				TotalMintAmount: 10_000_000, // Mint 10 tokens across the 10 days, or 1 per day
			},
			expectedMin: sdk.MustNewDecFromStr("0.99"), // 100% inflation
			expectedMax: sdk.MustNewDecFromStr("1.001"),
		},
		{
			name: "Good - Huge period - 10 years",
			minter: types.Minter{
				StartDate:       "2020-01-01", // 10 years of difference
				EndDate:         "2030-01-01",
				Denom:           "ukii",
				TotalMintAmount: 1_000_000_000, // Mint 100 tokens per year or 1_000 tokens total
			},
			expectedMin: sdk.MustNewDecFromStr("0.2730"), // 27.3% inflation (100/365)
			expectedMax: sdk.MustNewDecFromStr("0.2739"),
		},
		{
			name: "Good - Zero total mintable amount",
			minter: types.Minter{
				StartDate:       "2022-01-01",
				EndDate:         "2023-01-01",
				Denom:           "ukii",
				TotalMintAmount: 0, // No tokens to mint
			},
			expectedMin: sdk.ZeroDec(),
			expectedMax: sdk.ZeroDec(),
		},
		{
			name: "Bad - Mint duration is zero",
			minter: types.Minter{
				StartDate:       "2022-01-01",
				EndDate:         "2022-01-01",
				Denom:           "ukii",
				TotalMintAmount: 10_000_000,
			},
			errContains: "invalid minting duration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new cached context between tests
			cachedCtx, _ := ctx.CacheContext()

			// Call the function under test
			annualInflation, err := app.MintKeeper.GetAnnualInflationForMint(cachedCtx, &tc.minter)

			// Check for errors
			if tc.errContains == "" {
				require.NoError(t, err)
				// Check the annual inflation rate
				require.True(t, annualInflation.GTE(tc.expectedMin) && annualInflation.LTE(tc.expectedMax),
					"expected inflation between %s and %s, got %s", tc.expectedMin, tc.expectedMax, annualInflation)
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

// TestValidateNewMinter tests the validation of a new minter
func TestValidateNewMinter(t *testing.T) {
	// Initialize a new test app
	app, ctx := createTestApp(false)

	// Mint 100 tokens
	err := app.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100_000_000)),
	))
	require.NoError(t, err)

	// Prepare the test cases
	testCases := []struct {
		name        string
		minter      types.Minter
		malleate    func(sdk.Context)
		errContains string
	}{
		{
			name: "Good - Good mint state",
			minter: types.NewMinter(
				"2020-06-06",
				"2020-12-06", // 6 months
				"ukii",
				10_000_000, // 1 token per month, inflation of
			),
		},
		{
			name: "Bad - Above inflation rate",
			minter: types.NewMinter(
				"2020-06-06",
				"2020-12-06", // 6 months
				"ukii",
				15_000_000, // 1.5 tokens per month, inflation of 30%, above default 20%
			),
			errContains: "exceeds maximum allowed inflation rate",
		},
		{
			name: "Bad - Annual inflation rate fails",
			minter: types.NewMinter(
				"2020-06-06",
				"2020-06-06", // Same date
				"ukii",
				15_000_000, // 1.5 tokens per month, inflation of 30%, above default 20%
			),
			errContains: "invalid minting duration: must be greater than zero days",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new cached context between tests
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the system
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Now apply the tests
			err := app.MintKeeper.ValidateInflationRate(cachedCtx, &tc.minter)

			// Check for errors
			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
