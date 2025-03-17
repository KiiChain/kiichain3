package oracle_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	criptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/kiichain/kiichain/app"
	"github.com/kiichain/kiichain/precompiles/oracle"
	testkeeper "github.com/kiichain/kiichain/testutil/keeper"
	"github.com/kiichain/kiichain/x/evm/keeper"
	"github.com/kiichain/kiichain/x/evm/state"
	oracletypes "github.com/kiichain/kiichain/x/oracle/types"
	"github.com/kiichain/kiichain/x/oracle/utils"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestGetExchangeRates(t *testing.T) {
	// prepare env
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, types.Header{}).WithBlockHeight(2)
	evmKeeper := testApp.EvmKeeper
	oracleKeeper := testApp.OracleKeeper

	// create user account
	evm := setupEvmEnv(ctx, evmKeeper)

	// register exchange rates on the module
	rate := sdk.NewDec(1700)
	testApp.OracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, rate)

	// create precompiled
	precompile, err := oracle.NewPrecompile(oracleKeeper, &evmKeeper)
	require.NoError(t, err)

	executor := precompile.GetExecutor().(*oracle.PrecompileExecutor) // force to be an oracle executor
	query, err := precompile.ABI.MethodById(executor.GetExchangeRatesId)
	require.NoError(t, err)

	// perform a call to GetExchangeRates
	precompileRes, _, err := precompile.RunAndCalculateGas(
		evm,
		common.Address{},
		common.Address{},
		executor.GetExchangeRatesId,
		100000,
		nil, nil, true, false)
	require.NoError(t, err)

	// decode precompile response
	exchangeRates, err := query.Outputs.Unpack(precompileRes)
	require.NoError(t, err)

	// validate response
	require.Equal(t, 1, len(exchangeRates))

	// type assertion of the []interface{} response
	actualSlice, ok := exchangeRates[0].([]struct {
		Denom              string `json:"denom"`
		OracleExchangeRate struct {
			ExchangeRate        string   `json:"exchangeRate"`
			LastUpdate          string   `json:"lastUpdate"`
			LastUpdateTimestamp *big.Int `json:"lastUpdateTimestamp"`
		} `json:"oracleExchangeRate"`
	})
	require.True(t, ok)

	actual := actualSlice[0]
	require.Equal(t, utils.MicroAtomDenom, actual.Denom)
}

func TestGetOracleTwaps(t *testing.T) {
	// prepare env
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, types.Header{}).WithBlockHeight(2)
	evmKeeper := testApp.EvmKeeper
	oracleKeeper := testApp.OracleKeeper

	// Create test snapshots and insert on the module
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

	// set vote target on params
	params := oracletypes.DefaultParams()
	oracleKeeper.SetParams(ctx, params)
	for _, denom := range params.Whitelist {
		oracleKeeper.SetVoteTarget(ctx, denom.Name)
	}

	// setup sender and env
	evm := setupEvmEnv(ctx, evmKeeper)

	// create precompiled
	precompile, err := oracle.NewPrecompile(oracleKeeper, &evmKeeper)
	require.NoError(t, err)

	executor := precompile.GetExecutor().(*oracle.PrecompileExecutor)  // force to be an oracle executor
	query, err := precompile.ABI.MethodById(executor.GetOracleTwapsId) // create querier pointing to the function GetOracleTwaps
	require.NoError(t, err)

	// execute precompile
	args, err := query.Inputs.Pack(new(big.Int).SetUint64(3600)) // create the input arg
	require.NoError(t, err)
	precompileRes, _, err := precompile.RunAndCalculateGas(
		evm,
		common.Address{},
		common.Address{},
		append(executor.GetOracleTwapsId, args...),
		100000,
		nil, nil, true, false)
	require.Nil(t, err)

	twap, err := query.Outputs.Unpack(precompileRes)
	require.Nil(t, err)
	require.Equal(t, 1, len(twap))

	// type assertion of the []interface{} response
	actualSlice, ok := twap[0].([]struct {
		Denom           string   `json:"denom"`
		Twap            string   `json:"twap"`
		LookbackSeconds *big.Int `json:"lookbackSeconds"`
	})
	require.True(t, ok)

	require.Equal(t, utils.MicroEthDenom, actualSlice[0].Denom)
	require.Equal(t, sdk.NewDec(2).String(), actualSlice[0].Twap)
	require.Equal(t, utils.MicroKiiDenom, actualSlice[1].Denom)
	require.Equal(t, sdk.NewDec(1).String(), actualSlice[1].Twap)
}

func TestGetActives(t *testing.T) {
	// prepare env
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, types.Header{}).WithBlockHeight(2)
	evmKeeper := testApp.EvmKeeper
	oracleKeeper := testApp.OracleKeeper

	// setup sender and env
	evm := setupEvmEnv(ctx, evmKeeper)

	// Set Voting target
	voteTarget := map[string]oracletypes.Denom{
		utils.MicroKiiDenom:  {Name: utils.MicroKiiDenom},
		utils.MicroEthDenom:  {Name: utils.MicroEthDenom},
		utils.MicroUsdcDenom: {Name: utils.MicroUsdcDenom},
		utils.MicroAtomDenom: {Name: utils.MicroAtomDenom},
	}

	for denom := range voteTarget {
		oracleKeeper.SetVoteTarget(ctx, denom)
	}

	// create precompiled
	precompile, err := oracle.NewPrecompile(oracleKeeper, &evmKeeper)
	require.NoError(t, err)

	executor := precompile.GetExecutor().(*oracle.PrecompileExecutor) // force to be an oracle executor
	query, err := precompile.ABI.MethodById(executor.GetActivesId)    // create querier pointing to the function GetActives
	require.NoError(t, err)

	// execute precompile
	precompileRes, _, err := precompile.RunAndCalculateGas(
		evm,
		common.Address{},
		common.Address{},
		executor.GetActivesId,
		100000,
		nil, nil, true, false)
	require.Nil(t, err)

	actives, err := query.Outputs.Unpack(precompileRes)
	require.Nil(t, err)
	require.Equal(t, 1, len(actives))

	// type assertion of the []interface{} response
	activesSlice, ok := actives[0].([]string)
	require.True(t, ok)

	// validate response
	for _, denom := range activesSlice {
		_, ok := voteTarget[denom]
		require.True(t, ok)
	}
}

func TestGetPriceSnapshotHistory(t *testing.T) {
	// prepare env
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, types.Header{}).WithBlockHeight(2)
	evmKeeper := testApp.EvmKeeper
	oracleKeeper := testApp.OracleKeeper

	// setup sender and env
	evm := setupEvmEnv(ctx, evmKeeper)

	// Insert data on the module
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

	// test set and get snapshot data
	oracleKeeper.SetPriceSnapshot(ctx, snapshot1)
	oracleKeeper.SetPriceSnapshot(ctx, snapshot2)

	// create precompiled
	precompile, err := oracle.NewPrecompile(oracleKeeper, &evmKeeper)
	require.NoError(t, err)

	executor := precompile.GetExecutor().(*oracle.PrecompileExecutor)           // force to be an oracle executor
	query, err := precompile.ABI.MethodById(executor.GetPriceSnapshotHistoryId) // create querier
	require.NoError(t, err)

	// execute precompile
	precompileRes, _, err := precompile.RunAndCalculateGas(
		evm,
		common.Address{},
		common.Address{},
		executor.GetPriceSnapshotHistoryId,
		100000,
		nil, nil, true, false)
	require.Nil(t, err)

	history, err := query.Outputs.Unpack(precompileRes)
	require.Nil(t, err)
	require.Equal(t, 1, len(history))

	historySlice, ok := history[0].([]struct {
		SnapshotTimestamp  *big.Int `json:"snapshotTimestamp"`
		PriceSnapshotItems []struct {
			Denom              string `json:"denom"`
			OracleExchangeRate struct {
				ExchangeRate        string   `json:"exchangeRate"`
				LastUpdate          string   `json:"lastUpdate"`
				LastUpdateTimestamp *big.Int `json:"lastUpdateTimestamp"`
			} `json:"oracleExchangeRate"`
		} `json:"PriceSnapshotItems"`
	})
	require.True(t, ok)

	// data validation
	require.Equal(t, 2, len(historySlice))
	require.Equal(t, utils.MicroKiiDenom, historySlice[0].PriceSnapshotItems[0].Denom)
	require.Equal(t, utils.MicroKiiDenom, historySlice[1].PriceSnapshotItems[0].Denom)
}

func TestGetFeederDelegation(t *testing.T) {
	// prepare env
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, types.Header{}).WithBlockHeight(2)
	evmKeeper := testApp.EvmKeeper
	oracleKeeper := testApp.OracleKeeper

	// setup sender and env
	evm := setupEvmEnv(ctx, evmKeeper)

	// create validators
	privKey := secp256k1.GenPrivKey()
	valPub1 := privKey.PubKey()
	val1 := setupValidator(t, ctx, testApp, stakingtypes.Unbonded, valPub1)

	// create precompiled
	precompile, err := oracle.NewPrecompile(oracleKeeper, &evmKeeper)
	require.NoError(t, err)

	executor := precompile.GetExecutor().(*oracle.PrecompileExecutor)       // force to be an oracle executor
	query, err := precompile.ABI.MethodById(executor.GetFeederDelegationId) // create querier
	require.NoError(t, err)

	// execute precompile
	args, err := query.Inputs.Pack(val1.String()) // create the input arg
	precompileRes, _, err := precompile.RunAndCalculateGas(
		evm,
		common.Address{},
		common.Address{},
		append(executor.GetFeederDelegationId, args...),
		100000,
		nil, nil, true, false)
	require.Nil(t, err)

	feederByte, err := query.Outputs.Unpack(precompileRes)
	require.Nil(t, err)
	require.Equal(t, 1, len(feederByte))

	feederAddr, ok := feederByte[0].(string)
	require.True(t, ok)

	// data validation
	expectedAddress, _ := testkeeper.PrivateKeyToAddresses(privKey)
	require.Equal(t, expectedAddress.String(), feederAddr)
}

func TestGetVotePenaltyCounter(t *testing.T) {
	// prepare env
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, types.Header{}).WithBlockHeight(2)
	evmKeeper := testApp.EvmKeeper
	oracleKeeper := testApp.OracleKeeper

	// setup sender and env
	evm := setupEvmEnv(ctx, evmKeeper)

	// create validators
	privKey := secp256k1.GenPrivKey()
	valPub1 := privKey.PubKey()
	val1 := setupValidator(t, ctx, testApp, stakingtypes.Unbonded, valPub1)

	// create votes
	missCounter := uint64(10)
	abstainCounter := uint64(20)
	successCounter := uint64(30)
	oracleKeeper.SetVotePenaltyCounter(ctx, val1, missCounter, abstainCounter, successCounter) // Set the voting info

	// create precompiled
	precompile, err := oracle.NewPrecompile(oracleKeeper, &evmKeeper)
	require.NoError(t, err)

	executor := precompile.GetExecutor().(*oracle.PrecompileExecutor)         // force to be an oracle executor
	query, err := precompile.ABI.MethodById(executor.GetVotePenaltyCounterId) // create querier
	require.NoError(t, err)

	// execute precompile
	args, err := query.Inputs.Pack(val1.String()) // create the input arg
	precompileRes, _, err := precompile.RunAndCalculateGas(
		evm,
		common.Address{},
		common.Address{},
		append(executor.GetVotePenaltyCounterId, args...),
		100000,
		nil, nil, true, false)
	require.Nil(t, err)

	votePenaltyBytes, err := query.Outputs.Unpack(precompileRes)
	require.Nil(t, err)
	require.Equal(t, 1, len(votePenaltyBytes))

	// type assertion
	votePenalty, ok := votePenaltyBytes[0].(struct {
		MissCount    *big.Int `json:"missCount"`
		AbstainCount *big.Int `json:"abstainCount"`
		SuccessCount *big.Int `json:"successCount"`
	})
	require.True(t, ok)

	// data validation
	require.Equal(t, missCounter, votePenalty.MissCount.Uint64())
	require.Equal(t, abstainCounter, votePenalty.AbstainCount.Uint64())
	require.Equal(t, successCounter, votePenalty.SuccessCount.Uint64())
}

func setupEvmEnv(ctx sdk.Context, evmKeeper keeper.Keeper) *vm.EVM {
	privKey := testkeeper.MockPrivateKey()
	senderAddr, senderEVMAddr := testkeeper.PrivateKeyToAddresses(privKey)
	evmKeeper.SetAddressMapping(ctx, senderAddr, senderEVMAddr)
	statedb := state.NewDBImpl(ctx, &evmKeeper, true)
	evm := vm.EVM{
		StateDB:   statedb,
		TxContext: vm.TxContext{Origin: senderEVMAddr},
	}

	return &evm
}

func setupValidator(t *testing.T, ctx sdk.Context, a *app.App, bondStatus stakingtypes.BondStatus, valPub criptotypes.PubKey) sdk.ValAddress {
	valAddr := sdk.ValAddress(valPub.Address())
	bondDenom := a.StakingKeeper.GetParams(ctx).BondDenom
	selfBond := sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(100), Denom: bondDenom})

	err := a.BankKeeper.MintCoins(ctx, minttypes.ModuleName, selfBond)
	require.NoError(t, err)

	err = a.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdk.AccAddress(valAddr), selfBond)
	require.NoError(t, err)

	sh := teststaking.NewHelper(t, ctx, a.StakingKeeper)
	msg := sh.CreateValidatorMsg(valAddr, valPub, selfBond[0].Amount)
	sh.Handle(msg, true)

	val, found := a.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)

	val = val.UpdateStatus(bondStatus)
	a.StakingKeeper.SetValidator(ctx, val)

	consAddr, err := val.GetConsAddr()
	require.NoError(t, err)

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		ctx.BlockHeight(),
		0,
		time.Unix(0, 0),
		false,
		0,
	)
	a.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, signingInfo)

	return valAddr
}
