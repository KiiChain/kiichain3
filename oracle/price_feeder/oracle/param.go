package oracle

import (
	"context"
	"fmt"
	"time"

	oracletypes "github.com/kiichain/kiichain/x/oracle/types"
	"google.golang.org/grpc"
)

const (
	// paramsCacheInterval represents the amount of blocks
	// during which we will cache the oracle params.
	paramsCacheInterval = int64(200)
)

// ParamCache is used to cache oracle param data for
// an amount of blocks, defined by paramsCacheInterval.
type ParamCache struct {
	params           *oracletypes.Params
	lastUpdatedBlock int64
}

// Update updates the instance with the params information and update
// the last block analyzed
func (paramCache *ParamCache) Update(currentBlockHeight int64, params oracletypes.Params) {
	paramCache.lastUpdatedBlock = currentBlockHeight
	paramCache.params = &params
}

// IsOutdated checks whether or not the current
// param data was fetched in the last 200 blocks.
func (paramCache *ParamCache) IsOutdated(currentBlockHeight int64) bool {
	// check if ParamCache has the params
	if paramCache.params == nil {
		return true
	}

	// meanwhile the chain reaches the 200 blocks, the cached data is not outdated
	if currentBlockHeight < paramsCacheInterval {
		return false
	}

	// This is an edge case, which should never happen.
	// The current blockchain height is lower
	// than the last updated block, to fix we should
	// just update the cached params again.
	if currentBlockHeight < paramCache.lastUpdatedBlock {
		return true
	}

	return (currentBlockHeight - paramCache.lastUpdatedBlock) > paramsCacheInterval
}

// GetParamCache returns the last updated parameters of the oracle module
// if the current ParamCache is outdated, we will query it again.
func (o *Oracle) GetParamCache(ctx context.Context, currentBlockHeight int64) (oracletypes.Params, error) {
	// check if the param is outdated
	if !o.paramCache.IsOutdated(currentBlockHeight) {
		return *o.paramCache.params, nil
	}

	// query oracle module's params
	params, err := o.GetParams(ctx)
	if err != nil {
		return oracletypes.Params{}, err
	}

	o.checkWhitelist(params)

	// update params with the fetched info
	o.paramCache.Update(currentBlockHeight, params)
	return params, nil
}

// GetParams returns the current on-chain parameters of the x/oracle module.
func (o *Oracle) GetParams(ctx context.Context) (oracletypes.Params, error) {
	// create the connection with the blockchain
	grpcConn, err := grpc.Dial(
		o.oracleClient.GRPCEndpoint,
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithInsecure(),
		grpc.WithContextDialer(dialerFunc),
	)
	if err != nil {
		return oracletypes.Params{}, fmt.Errorf("failed to dial Cosmos gRPC service: %w", err)
	}

	defer grpcConn.Close()

	// create oracle query client
	queryClient := oracletypes.NewQueryClient(grpcConn)

	// create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// query oracle module's params
	queryResponse, err := queryClient.Params(ctx, &oracletypes.QueryParamsRequest{})
	if err != nil {
		return oracletypes.Params{}, fmt.Errorf("failed to get x/oracle params: %w", err)
	}

	// return params
	return *queryResponse.Params, nil
}

// checkWhitelist validates the denoms on the params' whitelist
// is on the oracle client chainDenomMapping
func (o *Oracle) checkWhitelist(params oracletypes.Params) {
	// iterate over the cached denom mapping
	chainDenomSet := make(map[string]struct{})
	for _, denom := range o.chainDenomMapping {
		chainDenomSet[denom] = struct{}{}
	}

	// iterate over the params's whitelist and validate every denom on
	// the whitelist is mapped on oracle client chainDenomMapping
	for _, denom := range params.Whitelist {
		_, ok := chainDenomSet[denom.Name]
		if !ok {
			o.logger.Warn().Str("denom", denom.Name).Msg("price missing for required denom")
		}
	}
}
