package wasmbinding

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosQuery "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/app"
	"github.com/kiichain/kiichain/wasmbinding"
	epochwasm "github.com/kiichain/kiichain/x/epoch/client/wasm"
	epochbinding "github.com/kiichain/kiichain/x/epoch/client/wasm/bindings"
	epochtypes "github.com/kiichain/kiichain/x/epoch/types"
	evmwasm "github.com/kiichain/kiichain/x/evm/client/wasm"
	oraclewasm "github.com/kiichain/kiichain/x/oracle/client/wasm"
	"github.com/kiichain/kiichain/x/oracle/types"
	"github.com/kiichain/kiichain/x/oracle/utils"
	tokenfactorywasm "github.com/kiichain/kiichain/x/tokenfactory/client/wasm"
	tokenfactorybinding "github.com/kiichain/kiichain/x/tokenfactory/client/wasm/bindings"
	tokenfactorytypes "github.com/kiichain/kiichain/x/tokenfactory/types"

	oraclebinding "github.com/kiichain/kiichain/x/oracle/client/wasm/bindings"
	oracletypes "github.com/kiichain/kiichain/x/oracle/types"
)

func SetupWasmbindingTest(t *testing.T) (*app.TestWrapper, func(ctx sdk.Context, request json.RawMessage) ([]byte, error)) {
	tm := time.Now().UTC()
	valPub := secp256k1.GenPrivKey().PubKey()

	testWrapper := app.NewTestWrapper(t, tm, valPub, false)

	eh := epochwasm.NewEpochWasmQueryHandler(&testWrapper.App.EpochKeeper)
	th := tokenfactorywasm.NewTokenFactoryWasmQueryHandler(&testWrapper.App.TokenFactoryKeeper)
	evmh := evmwasm.NewEVMQueryHandler(&testWrapper.App.EvmKeeper)
	oh := oraclewasm.NewOracleWasmQueryHandler(&testWrapper.App.OracleKeeper)
	qp := wasmbinding.NewQueryPlugin(eh, th, evmh, oh)
	return testWrapper, wasmbinding.CustomQuerier(qp)
}

func TestWasmGetEpoch(t *testing.T) {
	testWrapper, customQuerier := SetupWasmbindingTest(t)

	req := epochbinding.KiiEpochQuery{
		Epoch: &epochtypes.QueryEpochRequest{},
	}

	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{Route: wasmbinding.EpochRoute, QueryData: queryData}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	testWrapper.Ctx = testWrapper.Ctx.WithBlockHeight(45).WithBlockTime(time.Unix(12500, 0))
	testWrapper.App.EpochKeeper.SetEpoch(testWrapper.Ctx, epochtypes.Epoch{
		GenesisTime:           time.Unix(1000, 0).UTC(),
		EpochDuration:         time.Minute,
		CurrentEpoch:          uint64(69),
		CurrentEpochStartTime: time.Unix(12345, 0).UTC(),
		CurrentEpochHeight:    int64(40),
	})

	res, err := customQuerier(testWrapper.Ctx, rawQuery)
	require.NoError(t, err)

	var parsedRes epochtypes.QueryEpochResponse
	err = json.Unmarshal(res, &parsedRes)
	require.NoError(t, err)
	epoch := parsedRes.Epoch
	require.Equal(t, time.Unix(1000, 0).UTC(), epoch.GenesisTime)
	require.Equal(t, time.Minute, epoch.EpochDuration)
	require.Equal(t, uint64(69), epoch.CurrentEpoch)
	require.Equal(t, time.Unix(12345, 0).UTC(), epoch.CurrentEpochStartTime)
	require.Equal(t, int64(40), epoch.CurrentEpochHeight)
}

func TestWasmGetDenomAuthorityMetadata(t *testing.T) {
	testWrapper, customQuerier := SetupWasmbindingTest(t)

	denom := fmt.Sprintf("factory/%s/test", app.TestUser)
	testWrapper.Ctx = testWrapper.Ctx.WithBlockHeight(11).WithBlockTime(time.Unix(3600, 0))
	// Create denom
	testWrapper.App.TokenFactoryKeeper.CreateDenom(testWrapper.Ctx, app.TestUser, "test")
	authorityMetadata := tokenfactorytypes.DenomAuthorityMetadata{
		Admin: app.TestUser,
	}

	// Setup tfk query
	req := tokenfactorybinding.KiiTokenFactoryQuery{DenomAuthorityMetadata: &tokenfactorytypes.QueryDenomAuthorityMetadataRequest{Denom: denom}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{Route: wasmbinding.TokenFactoryRoute, QueryData: queryData}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	res, err := customQuerier(testWrapper.Ctx, rawQuery)
	require.NoError(t, err)

	var parsedRes tokenfactorytypes.QueryDenomAuthorityMetadataResponse
	err = json.Unmarshal(res, &parsedRes)
	require.NoError(t, err)
	require.Equal(t, tokenfactorytypes.QueryDenomAuthorityMetadataResponse{AuthorityMetadata: authorityMetadata}, parsedRes)
}

func TestWasmGetDenomsFromCreator(t *testing.T) {
	testWrapper, customQuerier := SetupWasmbindingTest(t)

	denom1 := fmt.Sprintf("factory/%s/test1", app.TestUser)
	denom2 := fmt.Sprintf("factory/%s/test2", app.TestUser)
	testWrapper.Ctx = testWrapper.Ctx.WithBlockHeight(11).WithBlockTime(time.Unix(3600, 0))

	// No denoms created initially
	req := tokenfactorybinding.KiiTokenFactoryQuery{DenomsFromCreator: &tokenfactorytypes.QueryDenomsFromCreatorRequest{Creator: app.TestUser}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{Route: wasmbinding.TokenFactoryRoute, QueryData: queryData}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	res, err := customQuerier(testWrapper.Ctx, rawQuery)
	require.NoError(t, err)

	var parsedRes tokenfactorytypes.QueryDenomsFromCreatorResponse
	err = json.Unmarshal(res, &parsedRes)
	require.NoError(t, err)
	require.Equal(t, tokenfactorytypes.QueryDenomsFromCreatorResponse{Denoms: nil, Pagination: &cosmosQuery.PageResponse{Total: 0}}, parsedRes)

	// Add first denom
	testWrapper.App.TokenFactoryKeeper.CreateDenom(testWrapper.Ctx, app.TestUser, "test1")

	res, err = customQuerier(testWrapper.Ctx, rawQuery)
	require.NoError(t, err)

	var parsedRes2 tokenfactorytypes.QueryDenomsFromCreatorResponse
	err = json.Unmarshal(res, &parsedRes2)
	require.NoError(t, err)
	require.Equal(t, tokenfactorytypes.QueryDenomsFromCreatorResponse{Denoms: []string{denom1}, Pagination: &cosmosQuery.PageResponse{Total: 1}}, parsedRes2)

	// Add second denom
	testWrapper.App.TokenFactoryKeeper.CreateDenom(testWrapper.Ctx, app.TestUser, "test2")

	res, err = customQuerier(testWrapper.Ctx, rawQuery)
	require.NoError(t, err)

	var parsedRes3 tokenfactorytypes.QueryDenomsFromCreatorResponse
	err = json.Unmarshal(res, &parsedRes3)
	require.NoError(t, err)
	require.Equal(t, tokenfactorytypes.QueryDenomsFromCreatorResponse{Denoms: []string{denom1, denom2}, Pagination: &cosmosQuery.PageResponse{Total: 2}}, parsedRes3)
}

func MockQueryPlugins() wasmkeeper.QueryPlugins {
	return wasmkeeper.QueryPlugins{
		Bank: func(ctx sdk.Context, request *wasmvmtypes.BankQuery) ([]byte, error) { return []byte{}, nil },
		IBC: func(ctx sdk.Context, caller sdk.AccAddress, request *wasmvmtypes.IBCQuery) ([]byte, error) {
			return []byte{}, nil
		},
		Custom: func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
			return []byte{}, nil
		},
		Stargate: func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) { return []byte{}, nil },
		Staking:  func(ctx sdk.Context, request *wasmvmtypes.StakingQuery) ([]byte, error) { return []byte{}, nil },
		Wasm:     func(ctx sdk.Context, request *wasmvmtypes.WasmQuery) ([]byte, error) { return []byte{}, nil },
	}
}

func TestOracleGetExchangeRates(t *testing.T) {
	// setup env
	testWrapper, customQuerier := SetupWasmbindingTest(t)
	ctx := testWrapper.Ctx

	// create query request
	req := oraclebinding.KiiOracleQuery{ExchangeRates: &types.QueryExchangeRatesRequest{}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{
		Route:     wasmbinding.OracleRoute,
		QueryData: queryData,
	}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	// execute query
	res, err := customQuerier(ctx, rawQuery)
	require.NoError(t, err)

	// process response
	parsedRes := &oracletypes.QueryExchangeRatesResponse{}
	err = json.Unmarshal(res, parsedRes)
	expectedResponse := &oracletypes.QueryExchangeRatesResponse{
		DenomOracleExchangeRate: oracletypes.DenomOracleExchangeRatePairs{},
	}

	// validate data
	require.NoError(t, err)
	require.Equal(t, expectedResponse, parsedRes)

	// simulate exchange rates on the module
	ctx = ctx.WithBlockHeight(11)
	testWrapper.App.OracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, sdk.NewDec(12))

	// execute query again
	res, err = customQuerier(ctx, rawQuery)
	require.NoError(t, err)

	// process response
	parsedRes2 := &oracletypes.QueryExchangeRatesResponse{}
	err = json.Unmarshal(res, parsedRes2)
	require.NoError(t, err)

	// validate data
	require.Equal(t, utils.MicroAtomDenom, parsedRes2.DenomOracleExchangeRate[0].Denom)
	require.Equal(t, sdk.NewDec(12), parsedRes2.DenomOracleExchangeRate[0].OracleExchangeRate.ExchangeRate)
}

func TestOracleGetOracleTwaps(t *testing.T) {
	// setup env
	testWrapper, customQuerier := SetupWasmbindingTest(t)
	oracleKeeper := testWrapper.App.OracleKeeper
	ctx := testWrapper.Ctx

	// create query request
	req := oraclebinding.KiiOracleQuery{OracleTwaps: &types.QueryTwapsRequest{LookbackSeconds: 200}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{
		Route:     wasmbinding.OracleRoute,
		QueryData: queryData,
	}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	// execute query (must fail because there is no snapshots to build a history)
	_, err = customQuerier(ctx, rawQuery)
	require.Error(t, err)

	// simulate snapshots to have history data
	ctx = ctx.WithBlockHeight(11).WithBlockTime(time.Unix(3600, 0))
	snapshotItem := oracletypes.NewPriceSnapshotItem(utils.MicroAtomDenom, oracletypes.OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(12),
		LastUpdate:          sdk.NewInt(10),
		LastUpdateTimestamp: ctx.BlockTime().Unix(),
	})
	snapshot := oracletypes.NewPriceSnapshot(3600, oracletypes.PriceSnapshotItems{snapshotItem})
	oracleKeeper.AddPriceSnapshot(ctx, snapshot)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)

	ctx = ctx.WithBlockHeight(12).WithBlockTime(time.Unix(3700, 0))

	// execute query again
	res, err := customQuerier(ctx, rawQuery)
	require.NoError(t, err)

	// process response
	parsedRes := &oracletypes.QueryTwapsResponse{}
	err = json.Unmarshal(res, parsedRes)
	require.NoError(t, err)

	// validate data
	require.Equal(t, utils.MicroAtomDenom, parsedRes.OracleTwap[0].Denom)
	require.Equal(t, int64(100), parsedRes.OracleTwap[0].LookbackSeconds)
}

func TestOracleGetActives(t *testing.T) {
	// setup env
	testWrapper, customQuerier := SetupWasmbindingTest(t)
	oracleKeeper := testWrapper.App.OracleKeeper
	ctx := testWrapper.Ctx

	// create query request
	req := oraclebinding.KiiOracleQuery{Actives: &types.QueryActivesRequest{}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{
		Route:     wasmbinding.OracleRoute,
		QueryData: queryData,
	}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	// Add actives to the blockchain (are gotten from the exchange rates uploaded)
	voteTarget := map[string]oracletypes.Denom{
		utils.MicroKiiDenom:  {Name: utils.MicroKiiDenom},
		utils.MicroEthDenom:  {Name: utils.MicroEthDenom},
		utils.MicroUsdcDenom: {Name: utils.MicroUsdcDenom},
		utils.MicroAtomDenom: {Name: utils.MicroAtomDenom},
	}

	for denom := range voteTarget {
		oracleKeeper.SetBaseExchangeRate(ctx, denom, sdk.NewDec(10))
	}

	// execute query
	res, err := customQuerier(ctx, rawQuery)
	require.NoError(t, err)

	// process response
	parsedRes := &oracletypes.QueryActivesResponse{}
	err = json.Unmarshal(res, parsedRes)
	require.NoError(t, err)

	// validate data
	for _, denom := range parsedRes.Actives {
		require.Equal(t, voteTarget[denom].Name, denom)
	}
}

func TestOracleGetPriceSnapshotHistory(t *testing.T) {
	// setup env
	testWrapper, customQuerier := SetupWasmbindingTest(t)
	oracleKeeper := testWrapper.App.OracleKeeper
	ctx := testWrapper.Ctx

	// create query request
	req := oraclebinding.KiiOracleQuery{PriceSnapshotHistory: &types.QueryPriceSnapshotHistoryRequest{}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{
		Route:     wasmbinding.OracleRoute,
		QueryData: queryData,
	}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	// Add snapshots
	exchangeRate1 := oracletypes.OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(1),
		LastUpdate:          sdk.NewInt(1),
		LastUpdateTimestamp: 1,
	}
	exchangeRate2 := oracletypes.OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(2),
		LastUpdate:          sdk.NewInt(2),
		LastUpdateTimestamp: 2,
	}
	snapshotItem1 := oracletypes.NewPriceSnapshotItem(utils.MicroKiiDenom, exchangeRate1)
	snapshotItem2 := oracletypes.NewPriceSnapshotItem(utils.MicroEthDenom, exchangeRate2)
	snapshot1 := oracletypes.NewPriceSnapshot(1, oracletypes.PriceSnapshotItems{snapshotItem1, snapshotItem1})
	snapshot2 := oracletypes.NewPriceSnapshot(2, oracletypes.PriceSnapshotItems{snapshotItem2, snapshotItem2})

	oracleKeeper.SetPriceSnapshot(ctx, snapshot1)
	oracleKeeper.SetPriceSnapshot(ctx, snapshot2)

	// execute query
	res, err := customQuerier(ctx, rawQuery)
	require.NoError(t, err)

	// process response
	parsedRes := &oracletypes.QueryPriceSnapshotHistoryResponse{}
	err = json.Unmarshal(res, parsedRes)
	require.NoError(t, err)

	// validate data
	require.Equal(t, utils.MicroKiiDenom, parsedRes.PriceSnapshot[0].PriceSnapshotItems[0].Denom)
	require.Equal(t, sdk.NewDec(1), parsedRes.PriceSnapshot[0].PriceSnapshotItems[0].OracleExchangeRate.ExchangeRate)
	require.Equal(t, utils.MicroKiiDenom, parsedRes.PriceSnapshot[0].PriceSnapshotItems[1].Denom)
	require.Equal(t, sdk.NewDec(1), parsedRes.PriceSnapshot[0].PriceSnapshotItems[1].OracleExchangeRate.ExchangeRate)
}

func TestOracleGetFeederDelegation(t *testing.T) {
	// setup env
	testWrapper, customQuerier := SetupWasmbindingTest(t)
	ctx := testWrapper.Ctx
	stakingKeeper := testWrapper.App.StakingKeeper

	// get validator and operator address
	val := stakingKeeper.GetValidators(ctx, 1)
	operatorAddr := val[0].GetOperator().String()

	// create query request
	req := oraclebinding.KiiOracleQuery{FeederDelegation: &types.QueryFeederDelegationRequest{ValidatorAddr: operatorAddr}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{
		Route:     wasmbinding.OracleRoute,
		QueryData: queryData,
	}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	// execute query
	res, err := customQuerier(ctx, rawQuery)
	require.NoError(t, err)

	// process response
	parsedRes := &oracletypes.QueryFeederDelegationResponse{}
	err = json.Unmarshal(res, parsedRes)
	require.NoError(t, err)

	// validate data (address response valid)
	_, err = sdk.AccAddressFromBech32(parsedRes.FeedAddr)
	require.NoError(t, err)
}

func TestOracleGetVotePenaltyCounter(t *testing.T) {
	// setup env
	testWrapper, customQuerier := SetupWasmbindingTest(t)
	ctx := testWrapper.Ctx
	oracleKeeper := testWrapper.App.OracleKeeper
	stakingKeeper := testWrapper.App.StakingKeeper

	// get validator and operator address
	val := stakingKeeper.GetValidators(ctx, 1)
	operatorAddr := val[0].GetOperator().String()

	// create query request
	req := oraclebinding.KiiOracleQuery{VotePenaltyCounter: &types.QueryVotePenaltyCounterRequest{ValidatorAddr: operatorAddr}}
	queryData, err := json.Marshal(req)
	require.NoError(t, err)
	query := wasmbinding.KiiQueryWrapper{
		Route:     wasmbinding.OracleRoute,
		QueryData: queryData,
	}

	rawQuery, err := json.Marshal(query)
	require.NoError(t, err)

	// Create voting penalty data on the module
	missCounter := uint64(10)
	abstainCounter := uint64(20)
	successCounter := uint64(30)
	oracleKeeper.SetVotePenaltyCounter(ctx, val[0].GetOperator(), missCounter, abstainCounter, successCounter) // Set the voting info

	// execute query
	res, err := customQuerier(ctx, rawQuery)
	require.NoError(t, err)

	// process response
	parsedRes := &oracletypes.QueryVotePenaltyCounterResponse{}
	err = json.Unmarshal(res, parsedRes)
	require.NoError(t, err)

	// validate data (address response valid)
	require.Equal(t, missCounter, parsedRes.VotePenaltyCounter.MissCount)
	require.Equal(t, abstainCounter, parsedRes.VotePenaltyCounter.AbstainCount)
	require.Equal(t, successCounter, parsedRes.VotePenaltyCounter.SuccessCount)
}
