package migrations_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"

	testkeeper "github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/kiichain/kiichain3/x/epoch/migrations"
)

// TestV2toV3Migration test the v2 to v3 migration
func TestV2toV3Migration(t *testing.T) {
	// Get the keeper and context
	k := testkeeper.EVMTestApp.EpochKeeper
	ctx := testkeeper.EVMTestApp.NewContext(false, tmtypes.Header{})

	// Run the migration
	migrations.V3MigrateStore(ctx, &k)

	// Check for panics
	require.NotPanics(t, func() { k.GetParams(ctx) })
}
