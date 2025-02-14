package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain3/x/epoch/types"
)

// TestEpochValidation tests the epoch validation process
func TestEpochValidation(t *testing.T) {
	// Get now and build a new epoch
	now := time.Now()

	// Prepare the test cases
	testCases := []struct {
		name        string
		epoch       *types.Epoch
		errContains string
	}{
		{
			name:  "Good - Default epoch",
			epoch: types.DefaultEpoch(),
		},
		{
			name: "Bad - Zero genesis time",
			epoch: types.NewEpoch(
				time.Time{},
				types.DefaultEpochDuration,
				0,
				now,
				0,
			),
			errContains: "epoch genesis time cannot be zero",
		},
		{
			name: "Bad - Zero duration",
			epoch: types.NewEpoch(
				now,
				0,
				0,
				now,
				0,
			),
			errContains: "epoch duration cannot be zero",
		},
		{
			name: "Bad - Giant epoch duration",
			epoch: types.NewEpoch(
				now,
				time.Hour*24, // One day
				0,
				now,
				0,
			),
			errContains: "epoch duration cannot exceed 3600.000000 seconds",
		},
		{
			name: "Bad - Genesis time after current epoch start time",
			epoch: types.NewEpoch(
				now.Add(time.Second), // One second after current epoch start time
				types.DefaultEpochDuration,
				0,
				now,
				0,
			),
			errContains: "epoch genesis time cannot be after epoch start time",
		},
		{
			name: "Bad - Current epoch heigh negative",
			epoch: types.NewEpoch(
				now,
				types.DefaultEpochDuration,
				0,
				now,
				-1,
			),
			errContains: "epoch current epoch height cannot be negative",
		},
	}

	// Run all the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.epoch.Validate()
			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
