package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/x/oracle/utils"
	"github.com/stretchr/testify/require"
)

func TestNewPriceSnapshotItem(t *testing.T) {
	rate := OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(11),
		LastUpdate:          sdk.NewInt(10),
		LastUpdateTimestamp: time.Now().Unix(),
	}

	// expected result
	expected := PriceSnapshotItem{
		Denom: utils.MicroAtomDenom,
		OracleExchangeRate: OracleExchangeRate{
			ExchangeRate:        sdk.NewDec(11),
			LastUpdate:          sdk.NewInt(10),
			LastUpdateTimestamp: time.Now().Unix(),
		},
	}

	// create snapshot item
	item := NewPriceSnapshotItem(utils.MicroAtomDenom, rate)

	// validate
	require.Equal(t, expected, item)
}

func TestNewPriceSnapshot(t *testing.T) {
	// price item 1
	rate1 := OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(11),
		LastUpdate:          sdk.NewInt(10),
		LastUpdateTimestamp: time.Now().Unix(),
	}
	item1 := NewPriceSnapshotItem(utils.MicroAtomDenom, rate1)

	// price item 2
	rate2 := OracleExchangeRate{
		ExchangeRate:        sdk.NewDec(22),
		LastUpdate:          sdk.NewInt(10),
		LastUpdateTimestamp: time.Now().Unix(),
	}
	item2 := NewPriceSnapshotItem(utils.MicroEthDenom, rate2)

	// create snapshot
	items := PriceSnapshotItems{item1, item2}
	snapshot := NewPriceSnapshot(12, items)

	// expected result
	expectedSnapshot := PriceSnapshot{
		SnapshotTimestamp: 12,
		PriceSnapshotItems: PriceSnapshotItems{
			PriceSnapshotItem{
				Denom: utils.MicroAtomDenom,
				OracleExchangeRate: OracleExchangeRate{
					ExchangeRate:        sdk.NewDec(11),
					LastUpdate:          sdk.NewInt(10),
					LastUpdateTimestamp: time.Now().Unix(),
				},
			},

			PriceSnapshotItem{
				Denom: utils.MicroEthDenom,
				OracleExchangeRate: OracleExchangeRate{
					ExchangeRate:        sdk.NewDec(22),
					LastUpdate:          sdk.NewInt(10),
					LastUpdateTimestamp: time.Now().Unix(),
				},
			},
		},
	}

	// validate
	require.Equal(t, expectedSnapshot, snapshot)
}
