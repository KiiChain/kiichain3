package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTokenReleaseSchedule(t *testing.T) {
	t.Parallel()
	t.Run("valid release schedule", func(t *testing.T) {
		validSchedule := []types.ScheduledTokenRelease{
			{
				StartDate:          "2023-01-01",
				EndDate:            "2023-01-31",
				TokenReleaseAmount: 1000,
			},
			{
				StartDate:          "2023-02-01",
				EndDate:            "2023-02-28",
				TokenReleaseAmount: 2000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(validSchedule)
		assert.Nil(t, err)
	})

	t.Run("invalid parameter type", func(t *testing.T) {
		invalidParam := "invalid"
		err := types.ValidateTokenReleaseSchedule(invalidParam)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid parameter type")
	})

	t.Run("invalid start date format", func(t *testing.T) {
		invalidStartDate := []types.ScheduledTokenRelease{
			{
				StartDate:          "invalid",
				EndDate:            "2023-01-31",
				TokenReleaseAmount: 1000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(invalidStartDate)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid start date format")
	})

	t.Run("invalid end date format", func(t *testing.T) {
		invalidEndDate := []types.ScheduledTokenRelease{
			{
				StartDate:          "2023-01-01",
				EndDate:            "invalid",
				TokenReleaseAmount: 1000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(invalidEndDate)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid end date format")
	})

	t.Run("start date not before end date", func(t *testing.T) {
		invalidDateOrder := []types.ScheduledTokenRelease{
			{
				StartDate:          "2023-01-31",
				EndDate:            "2023-01-01",
				TokenReleaseAmount: 1000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(invalidDateOrder)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "start date must be before end date")
	})

	t.Run("overlapping release period", func(t *testing.T) {
		overlappingPeriod := []types.ScheduledTokenRelease{
			{
				StartDate:          "2023-01-01",
				EndDate:            "2023-01-31",
				TokenReleaseAmount: 1000,
			},
			{
				StartDate:          "2023-01-15",
				EndDate:            "2023-01-31",
				TokenReleaseAmount: 2000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(overlappingPeriod)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "overlapping release period detected")
	})
	t.Run("non-overlapping periods with different order", func(t *testing.T) {
		nonOverlappingPeriods := []types.ScheduledTokenRelease{
			{
				StartDate:          "2023-03-01",
				EndDate:            "2023-03-31",
				TokenReleaseAmount: 3000,
			},
			{
				StartDate:          "2023-01-01",
				EndDate:            "2023-01-31",
				TokenReleaseAmount: 1000,
			},
			{
				StartDate:          "2023-02-01",
				EndDate:            "2023-02-28",
				TokenReleaseAmount: 2000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(nonOverlappingPeriods)
		assert.Nil(t, err)
	})

	t.Run("unsorted input with overlapping windows", func(t *testing.T) {
		unsortedOverlapping := []types.ScheduledTokenRelease{
			{
				StartDate:          "2023-03-01",
				EndDate:            "2023-03-31",
				TokenReleaseAmount: 3000,
			},
			{
				StartDate:          "2023-01-15",
				EndDate:            "2023-02-14",
				TokenReleaseAmount: 2000,
			},
			{
				StartDate:          "2023-01-01",
				EndDate:            "2023-01-31",
				TokenReleaseAmount: 1000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(unsortedOverlapping)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "overlapping release period detected")
	})

	t.Run("end date equals start date of next period is fine", func(t *testing.T) {
		endEqualsStart := []types.ScheduledTokenRelease{
			{
				StartDate:          "2023-01-01",
				EndDate:            "2023-01-31",
				TokenReleaseAmount: 1000,
			},
			{
				StartDate:          "2023-01-31",
				EndDate:            "2023-02-28",
				TokenReleaseAmount: 2000,
			},
			{
				StartDate:          "2023-02-28",
				EndDate:            "2023-03-31",
				TokenReleaseAmount: 3000,
			},
		}
		err := types.ValidateTokenReleaseSchedule(endEqualsStart)
		assert.Nil(t, err)
	})
}

// Validate params
func TestValidateParams(t *testing.T) {
	// The test cases
	testCases := []struct {
		name        string
		params      types.Params
		errContains string
	}{
		{
			name:   "Good - Default case",
			params: types.DefaultParams(),
		},
		{
			name:        "Good - Bad mint denom",
			params:      types.NewParams("test", nil, sdk.OneDec()),
			errContains: "mint denom must be the same as the default bond denom",
		},
		{
			name:        "Good - Bad max inflation",
			params:      types.NewParams("ukii", nil, sdk.NewDec(20)),
			errContains: "max inflation too large",
		},
		{
			name:        "Good - Bad max inflation (negative)",
			params:      types.NewParams("ukii", nil, sdk.NewDec(-1)),
			errContains: "max inflation cannot be negative",
		},
	}

	// Run the tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate the params
			err := tc.params.Validate()

			// Check for error
			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
