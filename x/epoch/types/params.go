package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// Parameter store keys
var (
	KeyMaxHooksGasAllowed = []byte("MaxHooksGasAllowed")
)

// Default values for params
const (
	DefaultMaxHooksGasAllowed = 10000000
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(maxHooksGasAllowed uint64) Params {
	return Params{
		MaxHooksGasAllowed: maxHooksGasAllowed,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxHooksGasAllowed)
}

// String implements the Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxHooksGasAllowed, &p.MaxHooksGasAllowed, validateMaxHooksGasAllowed),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	// Validate the max hooks gas allowed
	if err := validateMaxHooksGasAllowed(p.MaxHooksGasAllowed); err != nil {
		return err
	}

	return nil
}

// validateMaxHooksGasAllowed validates the max hooks allowed gas
func validateMaxHooksGasAllowed(i interface{}) error {
	// Check the interface
	maxAllowedGas, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// The value should not be zero
	if maxAllowedGas == 0 {
		return fmt.Errorf("epoch param max allowed gas can't be zero")
	}

	// We can't safely set a upper bound, so we return
	return nil
}
