package oracle

import (
	"context"
	"fmt"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc"
)

const (
	jailCacheIntervalBlocks = int64(50)
)

type JailCache struct {
	isJailed         bool
	lastUpdatedBlock int64
}

// Update updates the lastUpdatedBlock with the most recently block height analized
// also update the struct if the validator is jail
func (jailCache *JailCache) Update(currentBlockHeight int64, isJailed bool) {
	jailCache.lastUpdatedBlock = currentBlockHeight
	jailCache.isJailed = isJailed
}

// IsOutdated checks if the last analyzed block is further than 50 blocks
func (jailCache *JailCache) IsOutdated(currentBlockHeight int64) bool {
	if currentBlockHeight < jailCacheIntervalBlocks {
		return false
	}

	return (currentBlockHeight - jailCache.lastUpdatedBlock) > jailCacheIntervalBlocks
}

// GetCachedJailedState
func (o *Oracle) GetCachedJailedState(ctx context.Context, currentBlockHeight int64) (bool, error) {
	// check if the cached info is outdated (if no, return the cached data)
	if !o.jailCache.IsOutdated(currentBlockHeight) {
		return o.jailCache.isJailed, nil
	}

	// if the cached data is outdated fetch the validator's info
	isJailed, err := o.GetJailedState(ctx)
	if err != nil {
		return false, err
	}

	// update the cached info
	o.jailCache.Update(currentBlockHeight, isJailed)
	return isJailed, nil
}

// GetJailedState returns the current on-chain jailing state of the validator
func (o *Oracle) GetJailedState(ctx context.Context) (bool, error) {
	// create grpc connection with the blockchain
	grpcConn, err := grpc.Dial(
		o.oracleClient.GRPCEndpoint,
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithInsecure(),
		grpc.WithContextDialer(dialerFunc),
	)
	if err != nil {
		return false, fmt.Errorf("failed to dial Cosmos gRPC service: %w", err)
	}

	defer grpcConn.Close()

	// create staking query client
	queryClient := stakingtypes.NewQueryClient(grpcConn)

	// create context with 15s of timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)

	defer cancel() // cancel context when the function ends

	// query the validator information
	queryResponse, err := queryClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{ValidatorAddr: o.oracleClient.ValidatorAddrString})
	if err != nil {
		return false, fmt.Errorf("failed to get staking validator: %w", err)
	}

	// return the jail state of the validator
	return queryResponse.Validator.Jailed, nil
}
