package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/kiichain/kiichain3/x/oracle/types"
	"github.com/kiichain/kiichain3/x/oracle/utils"
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

func TestParams(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// test default params
	defaultParams := oracleKeeper.GetParams(ctx)
	oracleKeeper.SetParams(ctx, defaultParams)
	require.NotNil(t, defaultParams)

	// test custom params
	votePeriod := uint64(10)
	voteThreshold := sdk.NewDecWithPrec(33, 2) // 0.033
	rewardBand := sdk.NewDecWithPrec(1, 2)     // 0.01
	slashFraccion := sdk.NewDecWithPrec(1, 2)  // 0.01
	slashwindow := uint64(1000)
	minValPerWindow := sdk.NewDecWithPrec(1, 4) // 0.0001
	whiteList := types.DenomList{{Name: utils.MicroKiiDenom}, {Name: utils.MicroAtomDenom}}
	lookbackDuration := uint64(3600)

	params := types.Params{
		VotePeriod:        votePeriod,
		VoteThreshold:     voteThreshold,
		RewardBand:        rewardBand,
		Whitelist:         whiteList,
		SlashFraction:     slashFraccion,
		SlashWindow:       slashwindow,
		MinValidPerWindow: minValPerWindow,
		LookbackDuration:  lookbackDuration,
	}
	oracleKeeper.SetParams(ctx, params)

	storedParams := oracleKeeper.GetParams(ctx)
	require.NotNil(t, slashFraccion)
	require.Equal(t, params, storedParams)
}

func TestDelegationLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// ***** Get and set feeder delegator
	delegate := oracleKeeper.GetFeederDelegation(ctx, ValAddrs[0]) // supposed to received the same val addr
	require.Equal(t, Addrs[0], delegate)

	oracleKeeper.SetFeederDelegation(ctx, ValAddrs[0], Addrs[1]) // Delegate Val 0 -> Addr 1
	delegate = oracleKeeper.GetFeederDelegation(ctx, ValAddrs[0])
	require.Equal(t, Addrs[1], delegate)

	// ***** Iterate feeder delegator list
	var validators []sdk.ValAddress
	var delegates []sdk.AccAddress
	handler := func(valAddr sdk.ValAddress, delegatedFeeder sdk.AccAddress) bool {
		validators = append(validators, valAddr)
		delegates = append(delegates, delegatedFeeder)
		return false
	}
	oracleKeeper.IterateFeederDelegations(ctx, handler)

	// Validation
	require.Equal(t, 1, len(delegates))
	require.Equal(t, 1, len(validators))
	require.Equal(t, Addrs[1], delegates[0]) // Validator 0 delegate to -> Addr 1
}

func TestMissCounter(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// ***** Get default voting information
	counter := oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0

	// Validation (everything must be zero, I haven't add voting information yet)
	require.Equal(t, uint64(0), counter.MissCount)
	require.Equal(t, uint64(0), counter.AbstainCount)
	require.Equal(t, uint64(0), counter.SuccessCount)
	require.Equal(t, uint64(0), oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))

	// ***** Set an specific voting information
	missCounter := uint64(10)
	abstainCounter := uint64(20)
	successCounter := uint64(30)
	oracleKeeper.SetVotePenaltyCounter(ctx, ValAddrs[0], missCounter, abstainCounter, successCounter) // Set the voting info

	// Validation
	counter = oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0
	require.Equal(t, missCounter, counter.MissCount)
	require.Equal(t, abstainCounter, counter.AbstainCount)
	require.Equal(t, successCounter, counter.SuccessCount)
	require.Equal(t, missCounter, oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, abstainCounter, oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, successCounter, oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))

	// ***** Test delete voting info
	oracleKeeper.DeleteVotePanltyCounter(ctx, ValAddrs[0])

	// validation
	counter = oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0
	require.Equal(t, uint64(0), counter.MissCount)
	require.Equal(t, uint64(0), counter.AbstainCount)
	require.Equal(t, uint64(0), counter.SuccessCount)
	require.Equal(t, uint64(0), oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))

	// ***** Test increment function
	oracleKeeper.IncrementMissCount(ctx, ValAddrs[0])
	oracleKeeper.IncrementAbstainCount(ctx, ValAddrs[0])
	oracleKeeper.IncrementSuccessCount(ctx, ValAddrs[0])

	// validation
	counter = oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0
	require.Equal(t, uint64(1), counter.MissCount)
	require.Equal(t, uint64(1), counter.AbstainCount)
	require.Equal(t, uint64(1), counter.SuccessCount)
	require.Equal(t, uint64(1), oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(1), oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(1), oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))
}

func TestMissCounterIterate(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Set voting info
	missCounter := uint64(10)
	abstainCounter := uint64(20)
	successCounter := uint64(30)
	oracleKeeper.SetVotePenaltyCounter(ctx, ValAddrs[0], missCounter, abstainCounter, successCounter) // Set the voting info

	// The handler will iterate over
	handler := func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) bool {
		missCount := votePenaltyCounter.MissCount
		abstainCount := votePenaltyCounter.AbstainCount
		successCount := votePenaltyCounter.SuccessCount

		// validation
		require.Equal(t, missCounter, missCount)
		require.Equal(t, abstainCounter, abstainCount)
		require.Equal(t, successCounter, successCount)
		return true
	}

	oracleKeeper.IterateVotePenaltyCounters(ctx, handler)
}

func TestAggregateExchangeRateLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Create and aggregate exchange rate
	exchangeRate := types.ExchangeRateTuples{
		{Denom: "BTC/USD", ExchangeRate: sdk.NewDec(1)},
		{Denom: "ETH/USD", ExchangeRate: sdk.NewDec(2)},
		{Denom: "ATOM/USD", ExchangeRate: sdk.NewDec(3)},
	}
	exchangeRateVote, err := types.NewAggregateExchangeRateVote(exchangeRate, ValAddrs[0])
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[0], exchangeRateVote)
	require.NoError(t, err)

	// Get the aggregated exchange rate and validate
	gotExchangeRate, err := oracleKeeper.GetAggregateExchangeRateVote(ctx, ValAddrs[0])
	require.NoError(t, err)
	require.Equal(t, exchangeRate, gotExchangeRate.ExchangeRateTuples)
	require.Equal(t, ValAddrs[0].String(), gotExchangeRate.Voter)

	// Delete exchange rate
	oracleKeeper.DeleteAggregateExchangeRateVote(ctx, ValAddrs[0]) // delete exchange rate voting
	_, err = oracleKeeper.GetAggregateExchangeRateVote(ctx, ValAddrs[0])
	require.Error(t, err)

	// Create and aggregate invalid exchange rate
	exchangeRate = types.ExchangeRateTuples{
		{Denom: "BTC/USD", ExchangeRate: sdk.NewDec(0)},
		{Denom: "ETH/USD", ExchangeRate: sdk.NewDec(-1)},
		{Denom: "ATOM/USD", ExchangeRate: sdk.NewDec(2)},
	}
	_, err = types.NewAggregateExchangeRateVote(exchangeRate, ValAddrs[0])
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[0], exchangeRateVote)
	require.Error(t, err)
}
