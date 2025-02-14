package migrations_test

import (
	"testing"

	testkeeper "github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kiichain/kiichain3/x/mint/migrations"
)

// TestV3toV4Migration test the v3 to v4 migration
func TestV3toV4Migration(t *testing.T) {
	// Get the keeper and context
	k := testkeeper.EVMTestApp.MintKeeper
	ctx := testkeeper.EVMTestApp.NewContext(false, tmtypes.Header{})

	// Run the migration
	migrations.V4MigrateStore(ctx, &k)

	// Check for panics
	require.NotPanics(t, func() { k.GetParams(ctx) })
}
