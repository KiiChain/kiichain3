package artifacts

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/kiichain/kiichain/x/evm/artifacts/cw20"
	"github.com/kiichain/kiichain/x/evm/artifacts/cw721"
	"github.com/kiichain/kiichain/x/evm/artifacts/native"
)

func GetParsedABI(typ string) *abi.ABI {
	switch typ {
	case "native":
		return native.GetParsedABI()
	case "cw20":
		return cw20.GetParsedABI()
	case "cw721":
		return cw721.GetParsedABI()
	default:
		panic(fmt.Sprintf("unknown artifact type %s", typ))
	}
}

func GetBin(typ string) []byte {
	switch typ {
	case "native":
		return native.GetBin()
	case "cw20":
		return cw20.GetBin()
	case "cw721":
		return cw721.GetBin()
	default:
		panic(fmt.Sprintf("unknown artifact type %s", typ))
	}
}
