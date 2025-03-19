package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVoteTargets(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper

	// clear vote target
	oracleKeeper.ClearVoteTargets(input.Ctx)

	// set new expected targets
	expectedTargets := []string{"ukii", "ubtc", "ueth"}
	for _, target := range expectedTargets {
		oracleKeeper.SetVoteTarget(input.Ctx, target)
	}

	// validation
	targets := oracleKeeper.GetVoteTargets(input.Ctx)
	require.Equal(t, expectedTargets, targets)
}

func TestIsVoteTarget(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper

	// clear vote target
	oracleKeeper.ClearVoteTargets(input.Ctx)

	// set new expected targets and validate
	validTargets := []string{"ukii", "ubtc", "ueth"}
	for _, target := range validTargets {
		oracleKeeper.SetVoteTarget(input.Ctx, target)
		require.True(t, oracleKeeper.IsVoteTarget(input.Ctx, target))
	}
}
