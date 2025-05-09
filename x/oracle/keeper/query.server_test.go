package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/x/oracle/types"
	"github.com/kiichain/kiichain/x/oracle/utils"
	"github.com/stretchr/testify/require"
)

func TestQueryParams(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.Params(context, &types.QueryParamsRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, oracleKeeper.GetParams(ctx), *res.Params)
}

func TestQueryExchangeRate(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	rate := sdk.NewDec(12)
	oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, rate)

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.ExchangeRate(context, &types.QueryExchangeRateRequest{Denom: utils.MicroAtomDenom})

	// validation
	require.NoError(t, err)
	require.Equal(t, rate, res.OracleExchangeRate.ExchangeRate)
}
func TestQueryExchangeRates(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	rate := sdk.NewDec(12)
	oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, rate)
	oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroEthDenom, rate)

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.ExchangeRates(context, &types.QueryExchangeRatesRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, 2, len(res.DenomOracleExchangeRate))
}

func TestQueryActives(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	rate := sdk.NewDec(12)
	oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, rate)
	oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroEthDenom, rate)

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.Actives(context, &types.QueryActivesRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, 2, len(res.Actives))
	require.Equal(t, utils.MicroAtomDenom, res.Actives[0])
	require.Equal(t, utils.MicroEthDenom, res.Actives[1])
}

func TestQueryVoteTargets(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	oracleKeeper.DeleteVoteTargets(ctx)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroEthDenom)

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.VoteTargets(context, &types.QueryVoteTargetsRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, 2, len(res.VoteTargets))
	require.Equal(t, utils.MicroAtomDenom, res.VoteTargets[0])
	require.Equal(t, utils.MicroEthDenom, res.VoteTargets[1])
}

func TestQueryPriceSnapshotHistory(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	snapShot1 := types.NewPriceSnapshot(1, types.PriceSnapshotItems{
		types.NewPriceSnapshotItem(utils.MicroEthDenom, types.OracleExchangeRate{
			ExchangeRate: sdk.NewDec(11),
			LastUpdate:   sdk.NewInt(20),
		}),
		types.NewPriceSnapshotItem(utils.MicroAtomDenom, types.OracleExchangeRate{
			ExchangeRate: sdk.NewDec(12),
			LastUpdate:   sdk.NewInt(20),
		}),
	})

	snapShot2 := types.NewPriceSnapshot(2, types.PriceSnapshotItems{
		types.NewPriceSnapshotItem(utils.MicroEthDenom, types.OracleExchangeRate{
			ExchangeRate: sdk.NewDec(21),
			LastUpdate:   sdk.NewInt(30),
		}),
		types.NewPriceSnapshotItem(utils.MicroAtomDenom, types.OracleExchangeRate{
			ExchangeRate: sdk.NewDec(22),
			LastUpdate:   sdk.NewInt(30),
		}),
	})

	priceSnapshots := types.PriceSnapshots{snapShot1, snapShot2}

	oracleKeeper.SetPriceSnapshot(ctx, priceSnapshots[0])
	oracleKeeper.SetPriceSnapshot(ctx, priceSnapshots[1])

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.PriceSnapshotHistory(context, &types.QueryPriceSnapshotHistoryRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, priceSnapshots, res.PriceSnapshot)

}

func TestQueryTwaps(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	exchangeRate1 := types.OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(1),
		LastUpdate:          sdk.NewInt(1),
		LastUpdateTimestamp: 1,
	}
	exchangeRate2 := types.OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(2),
		LastUpdate:          sdk.NewInt(2),
		LastUpdateTimestamp: 2,
	}
	snapshotItem1 := types.NewPriceSnapshotItem(utils.MicroKiiDenom, exchangeRate1)
	snapshotItem2 := types.NewPriceSnapshotItem(utils.MicroEthDenom, exchangeRate2)
	snapshot1 := types.NewPriceSnapshot(1, types.PriceSnapshotItems{snapshotItem1, snapshotItem1})
	snapshot2 := types.NewPriceSnapshot(2, types.PriceSnapshotItems{snapshotItem2, snapshotItem2})

	oracleKeeper.SetPriceSnapshot(ctx, snapshot1)
	oracleKeeper.SetPriceSnapshot(ctx, snapshot2)

	// set vote target on params
	params := types.DefaultParams()
	oracleKeeper.SetParams(ctx, params)
	for _, denom := range params.Whitelist {
		oracleKeeper.SetVoteTarget(ctx, denom.Name)
	}

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.Twaps(context, &types.QueryTwapsRequest{LookbackSeconds: 3600})

	// validation
	require.NoError(t, err)
	require.Equal(t, utils.MicroEthDenom, res.OracleTwap[0].Denom)
	require.Equal(t, sdk.NewDec(2), res.OracleTwap[0].Twap)
}

func TestQueryFeederDelegation(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// delegate voting power
	oracleKeeper.SetFeederDelegation(ctx, ValAddrs[0], Addrs[0])

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.FeederDelegation(context, &types.QueryFeederDelegationRequest{ValidatorAddr: ValAddrs[0].String()})

	// validation
	require.NoError(t, err)
	require.Equal(t, Addrs[0].String(), res.FeedAddr)
}

func TestQueryVotePenaltyCounter(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// calculate the expected slashwindow
	missCounter := uint64(10)
	abstainCounter := uint64(20)
	successCounter := uint64(30)
	oracleKeeper.SetVotePenaltyCounter(ctx, ValAddrs[0], missCounter, abstainCounter, successCounter) // Set the voting info

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.VotePenaltyCounter(context, &types.QueryVotePenaltyCounterRequest{ValidatorAddr: ValAddrs[0].String()})

	// validation
	require.NoError(t, err)
	require.Equal(t, missCounter, res.VotePenaltyCounter.MissCount)
	require.Equal(t, abstainCounter, res.VotePenaltyCounter.AbstainCount)
	require.Equal(t, successCounter, res.VotePenaltyCounter.SuccessCount)
}

func TestQuerySlashWindow(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// calculate the expected slashwindow
	params := types.DefaultParams()
	expectedWindowProgress := (uint64(ctx.BlockHeight()) % params.SlashWindow) / params.VotePeriod

	// query params
	context := sdk.WrapSDKContext(ctx)
	res, err := querier.SlashWindow(context, &types.QuerySlashWindowRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, expectedWindowProgress, res.WindowProgress)
}
