package oracle

import (
	"embed"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	precommon "github.com/kiichain/kiichain/precompiles/common"
	"github.com/kiichain/kiichain/x/oracle/types"
)

// precompiled functions
const (
	GetExchangeRatesMethod        = "getExchangeRates"
	GetOracleTwapsMethod          = "getOracleTwaps"
	GetActivesMethod              = "getActives"
	GetPriceSnapshotHistoryMethod = "getPriceSnapshotHistory"
	GetFeederDelegationMethod     = "getFeederDelegation"
	GetVotePenaltyCounterMethod   = "getVotePenaltyCounter"
)

// precompiled address
const OracleAddress = "0x0000000000000000000000000000000000001008"

//go:embed abi.json
var f embed.FS

// PrecompileExecutor handles the precompile execution
type PrecompileExecutor struct {
	evmKeeper    precommon.EVMKeeper    // access point to the evm module
	oracleKeeper precommon.OracleKeeper // access point to the oracle module

	// functions to be registered
	GetExchangeRatesId        []byte
	GetOracleTwapsId          []byte
	GetActivesId              []byte
	GetPriceSnapshotHistoryId []byte
	GetFeederDelegationId     []byte
	GetVotePenaltyCounterId   []byte
}

// NewPrecompile registers the precompiled on the blockchain (this function is called on the app.go)
func NewPrecompile(oracleKeeper precommon.OracleKeeper, evmKeeper precommon.EVMKeeper) (*precommon.DynamicGasPrecompile, error) {
	// get abi with the contract interface
	newAbi := precommon.MustGetABI(f, "abi.json")

	// create precompile executor
	preExecutor := &PrecompileExecutor{
		evmKeeper:    evmKeeper,
		oracleKeeper: oracleKeeper,
	}

	// Save the solidity function's hash, based on the method name
	for name, method := range newAbi.Methods {
		switch name {
		case GetExchangeRatesMethod:
			preExecutor.GetExchangeRatesId = method.ID

		case GetOracleTwapsMethod:
			preExecutor.GetOracleTwapsId = method.ID

		case GetActivesMethod:
			preExecutor.GetActivesId = method.ID

		case GetPriceSnapshotHistoryMethod:
			preExecutor.GetPriceSnapshotHistoryId = method.ID

		case GetFeederDelegationMethod:
			preExecutor.GetFeederDelegationId = method.ID

		case GetVotePenaltyCounterMethod:
			preExecutor.GetVotePenaltyCounterId = method.ID
		}
	}

	return precommon.NewDynamicGasPrecompile(newAbi, preExecutor, common.HexToAddress(OracleAddress), "oracle"), nil
}

// EVMKeeper implements the interface DynamicGasPrecompileExecutor
// EVMKeeper returns the evm keeper
func (p PrecompileExecutor) EVMKeeper() precommon.EVMKeeper {
	return p.evmKeeper
}

// Execute implements the interface DynamicGasPrecompileExecutor
// Execute handles the contract call and execute the keeper function based on the contract function called
func (p PrecompileExecutor) Execute(ctx sdk.Context, method *abi.Method, caller common.Address, callingContract common.Address, args []interface{}, value *big.Int, readOnly bool, evm *vm.EVM, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	switch method.Name {
	case GetExchangeRatesMethod:
		return p.getExchangeRates(ctx, method, args, value)

	case GetOracleTwapsMethod:
		return p.getOracleTwaps(ctx, method, args, value)

	case GetActivesMethod:
		return p.getActives(ctx, method, args, value)

	case GetPriceSnapshotHistoryMethod:
		return p.getPriceSnapshotHistory(ctx, method, args, value)

	case GetFeederDelegationMethod:
		return p.getFeederDelegation(ctx, method, args, value)

	case GetVotePenaltyCounterMethod:
		return p.getVotePenaltyCounter(ctx, method, args, value)
	}
	return
}

// ******* FUNCTIONS CALLED BY THE PRECOMPILED

// OracleExchangeRate represents the exchange rate where the sdk.Dec is converted to string
type OracleExchangeRate struct {
	ExchangeRate        string
	LastUpdate          string
	LastUpdateTimestamp *big.Int
}

// DenomOracleExchangeRate represents the exchange rate by denom
type DenomOracleExchangeRate struct {
	Denom              string
	OracleExchangeRate OracleExchangeRate
}

// getExchangeRates returns the current exchange rates
func (p PrecompileExecutor) getExchangeRates(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// validate the function does not require payable
	if err := precommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}

	// validate the function does not receive args
	if err := precommon.ValidateArgsLength(args, 0); err != nil {
		return nil, 0, err
	}

	// Get exchange rates from oracle module
	exchangeRates := make([]DenomOracleExchangeRate, 0, 10)
	p.oracleKeeper.IterateBaseExchangeRates(ctx, func(denom string, exchangeRate types.OracleExchangeRate) bool {
		// parse the exchange rate from sdk.Dec to string
		rate := DenomOracleExchangeRate{
			Denom: denom,
			OracleExchangeRate: OracleExchangeRate{
				ExchangeRate:        exchangeRate.String(),
				LastUpdate:          exchangeRate.LastUpdate.String(),
				LastUpdateTimestamp: big.NewInt(exchangeRate.LastUpdateTimestamp),
			},
		}

		// store the exchange rates
		exchangeRates = append(exchangeRates, rate)
		return false
	})

	// convert from go struct to []byte data
	bz, err := method.Outputs.Pack(exchangeRates)
	if err != nil {
		return nil, 0, err

	}

	return bz, precommon.GetRemainingGas(ctx, p.evmKeeper), err
}

type OracleTwap struct {
	Denom           string
	Twap            string
	LookbackSeconds *big.Int
}

// getOracleTwaps calls the oracle keeper to calculate twaps within the lookback period
func (p PrecompileExecutor) getOracleTwaps(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// validate the function does not require payable
	if err := precommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}

	// validate the function receive only 1 arg
	if err := precommon.ValidateArgsLength(args, 1); err != nil {
		return nil, 0, err
	}

	// receive input arg
	lookbackSeconds := args[0].(*big.Int) // obligate the input is uint64

	// calculate twap
	twaps, err := p.oracleKeeper.CalculateTwaps(ctx, lookbackSeconds.Uint64())
	if err != nil {
		return nil, 0, err
	}

	// convert twaps to string
	stringTwaps := make([]OracleTwap, 0, len(twaps))
	for _, twap := range twaps {
		// convert twap to string
		stringTwap := OracleTwap{
			Denom:           twap.Denom,
			Twap:            twap.Twap.String(),
			LookbackSeconds: big.NewInt(twap.LookbackSeconds),
		}

		stringTwaps = append(stringTwaps, stringTwap)
	}

	// convert from go struct to []byte data
	bz, err := method.Outputs.Pack(stringTwaps)
	if err != nil {
		return nil, 0, err

	}

	return bz, precommon.GetRemainingGas(ctx, p.evmKeeper), err
}

// getActives returns the list of active assets
func (p PrecompileExecutor) getActives(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// validate the function does not require payable
	if err := precommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}

	// validate the function does not receive args
	if err := precommon.ValidateArgsLength(args, 0); err != nil {
		return nil, 0, err
	}

	// get the active assets
	denomsActive := []string{}
	p.oracleKeeper.IterateVoteTargets(ctx, func(denom string, denomInfo types.Denom) bool {
		denomsActive = append(denomsActive, denom)
		return false
	})

	// convert from go struct to []byte data
	bz, err := method.Outputs.Pack(denomsActive)
	if err != nil {
		return nil, 0, err

	}

	return bz, precommon.GetRemainingGas(ctx, p.evmKeeper), err
}

type PriceSnapshot struct {
	SnapshotTimestamp  *big.Int
	PriceSnapshotItems []DenomOracleExchangeRate
}

// getPriceSnapshotHistory returns the price history on string structs
func (p PrecompileExecutor) getPriceSnapshotHistory(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// validate the function does not require payable
	if err := precommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}

	// validate the function does not receive args
	if err := precommon.ValidateArgsLength(args, 0); err != nil {
		return nil, 0, err
	}

	// Get the snapshots available on the KVStore
	priceSnapshots := []PriceSnapshot{}
	snapshotItems := []DenomOracleExchangeRate{}

	// Get the snapshot list and convert to string
	p.oracleKeeper.IteratePriceSnapshots(ctx, func(snapshot types.PriceSnapshot) bool {
		// Iterate the snapshots by denom
		for _, item := range snapshot.PriceSnapshotItems {

			// convert the current rate to string
			stringRate := OracleExchangeRate{
				ExchangeRate:        item.OracleExchangeRate.ExchangeRate.String(),
				LastUpdate:          item.OracleExchangeRate.LastUpdate.String(),
				LastUpdateTimestamp: big.NewInt(item.OracleExchangeRate.LastUpdateTimestamp),
			}

			// create the string snapshot by denom
			snapshotItem := DenomOracleExchangeRate{
				Denom:              item.Denom,
				OracleExchangeRate: stringRate,
			}

			snapshotItems = append(snapshotItems, snapshotItem)
		}

		// create the struct with array of string snapshots by denom
		priceSnapshot := PriceSnapshot{
			SnapshotTimestamp:  big.NewInt(snapshot.SnapshotTimestamp),
			PriceSnapshotItems: snapshotItems,
		}

		priceSnapshots = append(priceSnapshots, priceSnapshot)
		return false
	})

	// convert from go struct to []byte data
	bz, err := method.Outputs.Pack(priceSnapshots)
	if err != nil {
		return nil, 0, err

	}

	return bz, precommon.GetRemainingGas(ctx, p.evmKeeper), err

}

// getFeederDelegation returns the delegation address based on the validator input arg
func (p PrecompileExecutor) getFeederDelegation(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// validate the function does not require payable
	if err := precommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}

	// validate the function receive only 1 arg
	if err := precommon.ValidateArgsLength(args, 1); err != nil {
		return nil, 0, err
	}

	// get the validator address from args
	valAddrString := args[0].(string) // obligate the string data type

	valAddr, err := sdk.ValAddressFromBech32(valAddrString)
	if err != nil {
		return nil, 0, err
	}

	// Get the delegator by the Validator address
	feederAddr := p.oracleKeeper.GetFeederDelegation(ctx, valAddr).String()

	// convert from go string to []byte data
	bz, err := method.Outputs.Pack(feederAddr)
	if err != nil {
		return nil, 0, err
	}

	return bz, precommon.GetRemainingGas(ctx, p.evmKeeper), err
}

type VotePenaltyCounter struct {
	MissCount    *big.Int
	AbstainCount *big.Int
	SuccessCount *big.Int
}

// getVotePenaltyCounter returns the penalty counter based on the validator input arg
func (p PrecompileExecutor) getVotePenaltyCounter(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// validate the function does not require payable
	if err := precommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}

	// validate the function receive only 1 arg
	if err := precommon.ValidateArgsLength(args, 1); err != nil {
		return nil, 0, err
	}

	// get the validator address from args
	valAddrString := args[0].(string) // obligate the string data type

	valAddr, err := sdk.ValAddressFromBech32(valAddrString)
	if err != nil {
		return nil, 0, err
	}

	// Get the penalty counters by the validator address
	missCount := p.oracleKeeper.GetMissCount(ctx, valAddr)
	abstainCount := p.oracleKeeper.GetAbstainCount(ctx, valAddr)
	successCount := p.oracleKeeper.GetSuccessCount(ctx, valAddr)

	votePenaltyCounter := VotePenaltyCounter{
		MissCount:    big.NewInt(int64(missCount)),
		AbstainCount: big.NewInt(int64(abstainCount)),
		SuccessCount: big.NewInt(int64(successCount)),
	}

	// convert from go struct to []byte data
	bz, err := method.Outputs.Pack(votePenaltyCounter)
	if err != nil {
		return nil, 0, err

	}

	return bz, precommon.GetRemainingGas(ctx, p.evmKeeper), err
}
