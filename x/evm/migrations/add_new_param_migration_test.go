package migrations_test

import (
	"testing"

	testkeeper "github.com/kiichain/kiichain/testutil/keeper"
	"github.com/kiichain/kiichain/x/evm/migrations"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestAddNewParamsAndSetAllToDefaults(t *testing.T) {
	k := testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.NewContext(false, tmtypes.Header{})
	migrations.AddNewParamsAndSetAllToDefaults(ctx, &k)
	require.NotPanics(t, func() { k.GetParams(ctx) })
}
