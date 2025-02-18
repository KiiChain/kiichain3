package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/kiichain/kiichain3/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestNewKeeper(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	encodingConfig := MakeEncodingConfig()
	cdc := encodingConfig.Marshaler

	// Create a new Keeper without causing a panic
	require.NotPanics(t, func() {
		NewKeeper(
			cdc,
			init.OracleKeeper.storeKey,
			init.OracleKeeper.memKey,
			init.OracleKeeper.paramSpace,
			init.AccountKeeper,
			init.BankKeeper,
			init.StakingKeeper,
			distTypes.ModuleName,
		)
	}, "NewKeeper should not panic if the Oracle module account is properly set")

	// Validate that paramSpace has a KeyTable after Keeper initialization
	oracleKeeper := NewKeeper(
		cdc,
		init.OracleKeeper.storeKey,
		init.OracleKeeper.memKey,
		init.OracleKeeper.paramSpace,
		init.AccountKeeper,
		init.BankKeeper,
		init.StakingKeeper,
		distTypes.ModuleName,
	)

	require.True(t, oracleKeeper.paramSpace.HasKeyTable(), "paramSpace in the Keeper should have a KeyTable")
}

func TestExchangeRateLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Exchange rates to be stored
	const BTC_USD = "BTC/USD"
	const ETH_USD = "ETC/USD"
	const ATOM_USD = "ATOM/USD"

	btcUsdExchangeRate := sdk.NewDecWithPrec(100, int64(OracleDecPrecision)).MulInt64(1e6)
	ethUsdExchangeRate := sdk.NewDecWithPrec(200, int64(OracleDecPrecision)).MulInt64(1e6)
	atomUsdExchangeRate := sdk.NewDecWithPrec(300, int64(OracleDecPrecision)).MulInt64(1e6)

	// ***** First exchange rate insertion
	oracleKeeper.SetBaseExchangeRate(ctx, BTC_USD, btcUsdExchangeRate)               // Set exchange rates on KVStore
	btcUsdRate, lastUpdate, _, err := oracleKeeper.GetBaseExchangeRate(ctx, BTC_USD) // Get exchange rate from KVStore
	require.NoError(t, err, "Expected no error getting BTC/USD exchange rate")
	require.Equal(t, btcUsdExchangeRate, btcUsdRate, "Expected got the same exchange rate as ")
	require.Equal(t, sdk.ZeroInt(), lastUpdate) // There is no previous updates

	// simulate time pass
	ctx = ctx.WithBlockHeight(3) // Update block height
	ts := time.Now()
	ctx = ctx.WithBlockTime(ts) // Update block timestamp

	// ***** Second exchange rate insertion
	oracleKeeper.SetBaseExchangeRate(ctx, ETH_USD, ethUsdExchangeRate)                                 // Set exchange rates on KVStore
	ethUsdRate, lastUpdate, lastUpdateTimestamp, err := oracleKeeper.GetBaseExchangeRate(ctx, ETH_USD) // Get exchange rate from KVStore
	require.NoError(t, err)
	require.Equal(t, ethUsdExchangeRate, ethUsdRate)
	require.Equal(t, sdk.NewInt(3), lastUpdate)
	require.Equal(t, ts.UnixMilli(), lastUpdateTimestamp)

	// simulate time pass
	ctx = ctx.WithBlockHeight(15) // Update block height
	newTime := ts.Add(time.Hour)
	ctx = ctx.WithBlockTime(newTime) // Update block timestamp

	// ***** Third exchange rate insertion (using events)
	oracleKeeper.SetBaseExchangeRateWithEvent(ctx, ATOM_USD, atomUsdExchangeRate)                        // Set exchange rates on KVStore
	atomUsdRate, lastUpdate, lastUpdateTimestamp, err := oracleKeeper.GetBaseExchangeRate(ctx, ATOM_USD) // Get exchange rate from KVStore

	// Create the event validation function
	eventValidation := func() bool {
		// Expected event
		expectedEvent := sdk.NewEvent(
			types.EventTypeExchangeRateUpdate,
			sdk.NewAttribute(types.AttributeKeyDenom, ATOM_USD),
			sdk.NewAttribute(types.AttributeKeyExchangeRate, atomUsdExchangeRate.String()))

		// Read the current events
		events := ctx.EventManager().Events()
		for _, event := range events {
			if event.Type != expectedEvent.Type { // Search the expected event
				continue
			}

			// Iterate over the event
			for i, attr := range event.Attributes {
				if attr.Index != expectedEvent.Attributes[i].Index {
					return false
				}

				if attr.Key != expectedEvent.Attributes[i].Key {
					return false
				}

				if attr.Value != expectedEvent.Attributes[i].Value {
					return false
				}
			}
			return true
		}
		return false
	}

	// Validations
	require.NoError(t, err)
	require.Equal(t, atomUsdExchangeRate, atomUsdRate)
	require.Equal(t, sdk.NewInt(15), lastUpdate)
	require.Equal(t, newTime.UnixMilli(), lastUpdateTimestamp)
	require.True(t, eventValidation())

	// ***** First exchange rate elimination
	oracleKeeper.DeleteBaseExchangeRate(ctx, BTC_USD)
	_, _, _, err = oracleKeeper.GetBaseExchangeRate(ctx, BTC_USD)
	require.Error(t, err) // Validate error

	// test iteration function
	exchangeRateAmount := 0
	iterationHandler := func(denom string, exchangeRate types.OracleExchangeRate) bool {
		exchangeRateAmount++
		return false
	}

	oracleKeeper.IterateBaseExchangeRates(ctx, iterationHandler)
	require.Equal(t, 2, exchangeRateAmount) // verify that iterate over all exchange rates elements
}


