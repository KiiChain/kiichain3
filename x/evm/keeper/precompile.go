package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kiichain/kiichain/precompiles/bank"
	"github.com/kiichain/kiichain/precompiles/gov"
	"github.com/kiichain/kiichain/precompiles/staking"
	"github.com/kiichain/kiichain/precompiles/wasmd"
)

// add any payable precompiles here
// these will suppress transfer events to/from the precompile address
var payablePrecompiles = map[string]struct{}{
	bank.BankAddress:       {},
	staking.StakingAddress: {},
	gov.GovAddress:         {},
	wasmd.WasmdAddress:     {},
}

func IsPayablePrecompile(addr *common.Address) bool {
	if addr == nil {
		return false
	}
	_, ok := payablePrecompiles[addr.Hex()]
	return ok
}
