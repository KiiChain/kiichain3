package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kiichain/kiichain3/x/oracle/types"
	"github.com/kiichain/kiichain3/x/oracle/utils"
	"github.com/stretchr/testify/require"
)

func TestOrganizeBallotByDenom(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	stakingKeeper := init.StakingKeeper
	ctx := init.Ctx

	// Create handlers
	stakingHandler := staking.NewHandler(stakingKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	val0 := NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakingAmount)
	val1 := NewTestMsgCreateValidator(ValAddrs[1], ValPubKeys[1], stakingAmount)

	// Register validators
	_, err := stakingHandler(ctx, val0)
	require.NoError(t, err)
	_, err = stakingHandler(ctx, val1)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	staking.EndBlocker(ctx, stakingKeeper)

	// Simulate aggregation exchange rate process
	exchangeRate1 := types.ExchangeRateTuples{
		{Denom: utils.MicroAtomDenom, ExchangeRate: sdk.NewDec(1)},
		{Denom: utils.MicroEthDenom, ExchangeRate: sdk.NewDec(2)},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: sdk.NewDec(3)},
		{Denom: utils.MicroKiiDenom, ExchangeRate: sdk.NewDec(4)},
	}

	exchangeRate2 := types.ExchangeRateTuples{
		{Denom: utils.MicroAtomDenom, ExchangeRate: sdk.NewDec(1)},
		{Denom: utils.MicroEthDenom, ExchangeRate: sdk.NewDec(2)},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: sdk.NewDec(3)},
		{Denom: utils.MicroKiiDenom, ExchangeRate: sdk.NewDec(4)},
	}

	exchangeRateVote1, err := types.NewAggregateExchangeRateVote(exchangeRate1, ValAddrs[0]) // Aggregate rate tuples from Val0
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[0], exchangeRateVote1)
	require.NoError(t, err)

	exchangeRateVote2, err := types.NewAggregateExchangeRateVote(exchangeRate2, ValAddrs[1]) // Aggregate rate tuples from Val1
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[1], exchangeRateVote2)
	require.NoError(t, err)

	// Get claim map
	validatorClaimMap := make(map[string]types.Claim)
	powerReduction := stakingKeeper.PowerReduction(ctx)

	iterator := stakingKeeper.ValidatorsPowerStoreIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		valAddr := sdk.ValAddress(iterator.Value())        // Get validator address
		validator := stakingKeeper.Validator(ctx, valAddr) // get validator by address

		valPower := validator.GetConsensusPower(powerReduction)
		operator := validator.GetOperator()
		claim := types.NewClaim(valPower, 0, 0, false, operator)

		validatorClaimMap[operator.String()] = claim // Assign the validator on the list to receive
	}

	// Create expected result (with denom organized alphabetically)
	uatomBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroAtomDenom, ExchangeRate: sdk.NewDec(1), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: sdk.NewDec(1), Power: int64(10), Voter: ValAddrs[1]},
	}

	uethBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroEthDenom, ExchangeRate: sdk.NewDec(2), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroEthDenom, ExchangeRate: sdk.NewDec(2), Power: int64(10), Voter: ValAddrs[1]},
	}

	uusdcBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroUsdcDenom, ExchangeRate: sdk.NewDec(3), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: sdk.NewDec(3), Power: int64(10), Voter: ValAddrs[1]},
	}

	ukiiBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroKiiDenom, ExchangeRate: sdk.NewDec(4), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroKiiDenom, ExchangeRate: sdk.NewDec(4), Power: int64(10), Voter: ValAddrs[1]},
	}

	reference := map[string]types.ExchangeRateBallot{
		utils.MicroAtomDenom: uatomBallot,
		utils.MicroEthDenom:  uethBallot,
		utils.MicroUsdcDenom: uusdcBallot,
		utils.MicroKiiDenom:  ukiiBallot,
	}

	// Call function
	denomBallot := oracleKeeper.OrganizeBallotByDenom(ctx, validatorClaimMap)
	require.Equal(t, reference, denomBallot)
}
