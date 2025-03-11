package keeper

import (
	"github.com/kiichain/kiichain/x/epoch/types"
)

var _ types.QueryServer = Keeper{}
