package types

import (
	fmt "fmt"
	"time"
)

const (
	// Default epoch params
	DefaultEpochDuration = time.Minute
	DefaultCurrentEpoch  = 0
	DefaultEpochHeight   = 0

	// MaxEpochDuration is the max epoch duration of one hour
	MaxEpochDuration = time.Hour
)

// NewEpoch creates a new Epoch instance
func NewEpoch(
	genesisTime time.Time,
	epochDuration time.Duration,
	currentEpoch uint64,
	currentEpochStartTime time.Time,
	currentEpochHeight int64,
) *Epoch {
	return &Epoch{
		GenesisTime:           genesisTime,
		EpochDuration:         epochDuration,
		CurrentEpoch:          currentEpoch,
		CurrentEpochStartTime: currentEpochStartTime,
		CurrentEpochHeight:    currentEpochHeight,
	}
}

// DefaultParams returns a default set of parameters
func DefaultEpoch() *Epoch {
	// Get now and build a new epoch
	now := time.Now().UTC()

	// Return the epoch
	return NewEpoch(
		now,
		DefaultEpochDuration,
		DefaultCurrentEpoch,
		now,
		DefaultEpochHeight,
	)
}

// Validate validates the epoch
func (e *Epoch) Validate() error {
	// Check if genesis time is zero
	if e.GetGenesisTime().IsZero() {
		return fmt.Errorf("epoch genesis time cannot be zero")
	}

	// Get the epoch duration to be used on other validators
	epochDuration := e.GetEpochDuration().Seconds()

	// Check the epoch duration
	if epochDuration == 0 {
		return fmt.Errorf("epoch duration cannot be zero")
	}
	if epochDuration > MaxEpochDuration.Seconds() {
		return fmt.Errorf("epoch duration cannot exceed %f seconds (got %f)", MaxEpochDuration.Seconds(), epochDuration)
	}

	// Check genesis time
	if e.GetGenesisTime().After(e.GetCurrentEpochStartTime()) {
		return fmt.Errorf("epoch genesis time cannot be after epoch start time")
	}

	// Check current epoch height
	if e.GetCurrentEpochHeight() < 0 {
		return fmt.Errorf("epoch current epoch height cannot be negative")
	}

	return nil
}
