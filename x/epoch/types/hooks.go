package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain3/utils"
)

// EpochHooks are is the interface for the Epoch Hooks
type EpochHooks interface {
	// AfterEpochEnd defines the first block whose timestamp is after the duration
	// is counted as the end of the epoch.
	AfterEpochEnd(ctx sdk.Context, epoch Epoch)
	// BeforeEpochStart defines the new epoch is next block of epoch EndBlock.
	BeforeEpochStart(ctx sdk.Context, epoch Epoch)
}

// MultiEpochHooks is the set of epoch hooks
type MultiEpochHooks []EpochHooks

// NewMultiEpochHooks returns a new MultiEpochHooks
func NewMultiEpochHooks(hooks ...EpochHooks) MultiEpochHooks {
	return hooks
}

// AfterEpochEnd is called when epoch is going to be ended, epochNumber is the
// number of epoch that is ending.
func (h MultiEpochHooks) AfterEpochEnd(ctx sdk.Context, epoch Epoch, maxHooksGasAllowed sdk.Gas) {
	// List of hooks to be executed
	hooksFns := make([]func(sdk.Context, Epoch), len(h))

	// Iterate the hooks
	for i, hook := range h {
		hooksFns[i] = hook.AfterEpochEnd
	}

	// Execute with gas limit
	executeHookWithGasLimit(ctx, hooksFns, epoch, maxHooksGasAllowed)
}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is
// the number of epoch that is starting.
func (h MultiEpochHooks) BeforeEpochStart(ctx sdk.Context, epoch Epoch, maxHooksGasAllowed sdk.Gas) {
	// List of hooks to be executed
	hooksFns := make([]func(sdk.Context, Epoch), len(h))

	// Run the loop and check for gas usage
	for i, hook := range h {
		hooksFns[i] = hook.BeforeEpochStart
	}

	// Execute with gas limit
	executeHookWithGasLimit(ctx, hooksFns, epoch, maxHooksGasAllowed)
}

// executeHookWithGasLimit execute hooks with a gas limit
func executeHookWithGasLimit(
	ctx sdk.Context,
	hooks []func(sdk.Context, Epoch),
	epoch Epoch,
	maxHooksGasAllowed sdk.Gas,
) {
	// Start the gas metric
	limitedCtx := ctx.WithGasMeter(sdk.NewGasMeter(maxHooksGasAllowed, 1, 1))

	// Iterate the hooks
	for _, hookFn := range hooks {
		// Execute the hook
		panicCatchingEpochHook(limitedCtx, hookFn, epoch)
	}
}

// panicCatchingEpochHook catch panics from hooks execution
func panicCatchingEpochHook(ctx sdk.Context, hookFn func(sdk.Context, Epoch), epoch Epoch) {
	defer utils.PanicHandler(func(r any) {
		utils.LogPanicCallback(ctx, r)
	})()

	// cache the context and only write if no panic (which is caught above)
	cacheCtx, write := ctx.CacheContext()
	hookFn(cacheCtx, epoch)
	write()
}
