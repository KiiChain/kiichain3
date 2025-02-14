package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain3/x/epoch/types"
)

// TestParamsValidation validation of parameters
func TestParamsValidation(t *testing.T) {
	// Prepare the test cases
	testCases := []struct {
		name        string
		epoch       types.Params
		errContains string
	}{
		{
			name:  "Good - Default epoch",
			epoch: types.DefaultParams(),
		},
		{
			name:        "Bad - Max gas is zero",
			epoch:       types.NewParams(0),
			errContains: "epoch param max allowed gas can't be zero",
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
