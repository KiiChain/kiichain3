package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/app"
	"github.com/kiichain/kiichain/x/epoch/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type mockEpochHooks struct {
	afterEpochEndCalled    bool
	beforeEpochStartCalled bool
}

func (h *mockEpochHooks) AfterEpochEnd(ctx sdk.Context, _ types.Epoch, _ sdk.Gas) {
	h.afterEpochEndCalled = true
}

func (h *mockEpochHooks) BeforeEpochStart(ctx sdk.Context, _ types.Epoch, _ sdk.Gas) {
	h.beforeEpochStartCalled = true
}

type multiHooksMock struct {
	afterEpochEndCalled    bool
	beforeEpochStartCalled bool
	gasToConsume           sdk.Gas
}

func (h *multiHooksMock) AfterEpochEnd(ctx sdk.Context, _ types.Epoch) {
	ctx.GasMeter().ConsumeGas(h.gasToConsume, "mock hook")

	h.afterEpochEndCalled = true
}

func (h *multiHooksMock) BeforeEpochStart(ctx sdk.Context, _ types.Epoch) {
	ctx.GasMeter().ConsumeGas(h.gasToConsume, "mock hook")

	h.beforeEpochStartCalled = true
}

func TestKeeperHooks(t *testing.T) {
	// Start a app
	app := app.Setup(false, false) // Your setup function here
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Get the keeper
	k := app.EpochKeeper
	hooks := &mockEpochHooks{}

	// Can't set the same hook twice
	require.Panics(t, func() {
		k.SetHooks(hooks)
	})

	// For the tests use the mock hook
	k.UnsafeSetHooks(hooks)
	epoch := types.Epoch{} // setup epoch as required

	k.AfterEpochEnd(ctx, epoch)
	require.True(t, hooks.afterEpochEndCalled)

	hooks.afterEpochEndCalled = false // reset for the next test

	k.BeforeEpochStart(ctx, epoch)
	require.True(t, hooks.beforeEpochStartCalled)
}

// TestHooksGasLimit test the epoch hooks gas limit
func TestHooksGasLimit(t *testing.T) {
	// Start a new app
	app := app.Setup(false, false) // Your setup function here
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Define an epoch
	currentTime := time.Now().UTC()
	epochIn := types.Epoch{
		CurrentEpochStartTime: currentTime,
		CurrentEpochHeight:    100,
	}

	// Set the epoch
	app.EpochKeeper.SetEpoch(ctx, epochIn)

	// Set the gas limit
	app.EpochKeeper.SetParams(ctx, types.Params{MaxHooksGasAllowed: 200})

	// Set the mock hooks
	hook1 := &multiHooksMock{gasToConsume: 150}
	hook2 := &multiHooksMock{gasToConsume: 150}
	app.EpochKeeper.UnsafeSetHooks(types.MultiEpochHooks{
		hook1,
		hook2,
	})

	// Set the hooks on the keeper

	// Run the epoch
	app.EpochKeeper.BeforeEpochStart(ctx, epochIn)
	app.EpochKeeper.AfterEpochEnd(ctx, epochIn)

	// Check the execution
	// Hook 1 should run
	require.True(t, hook1.beforeEpochStartCalled)
	require.True(t, hook1.afterEpochEndCalled)

	// Hook 2 should fail
	require.False(t, hook2.beforeEpochStartCalled)
	require.False(t, hook2.afterEpochEndCalled)

	// Now we bump the gas
	app.EpochKeeper.SetParams(ctx, types.Params{MaxHooksGasAllowed: 2000})

	// Set the mock hooks
	hook1 = &multiHooksMock{gasToConsume: 750}
	hook2 = &multiHooksMock{gasToConsume: 750}
	app.EpochKeeper.UnsafeSetHooks(types.MultiEpochHooks{
		hook1,
		hook2,
	})

	// Run the hooks
	app.EpochKeeper.BeforeEpochStart(ctx, epochIn)
	app.EpochKeeper.AfterEpochEnd(ctx, epochIn)

	// Now 1 and 2 should have passed
	require.True(t, hook1.beforeEpochStartCalled)
	require.True(t, hook1.afterEpochEndCalled)
	require.True(t, hook2.beforeEpochStartCalled)
	require.True(t, hook2.afterEpochEndCalled)
}
