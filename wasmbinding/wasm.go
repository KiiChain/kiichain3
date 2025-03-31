package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	aclkeeper "github.com/cosmos/cosmos-sdk/x/accesscontrol/keeper"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	epochwasm "github.com/kiichain/kiichain/x/epoch/client/wasm"
	epochkeeper "github.com/kiichain/kiichain/x/epoch/keeper"
	evmwasm "github.com/kiichain/kiichain/x/evm/client/wasm"
	evmkeeper "github.com/kiichain/kiichain/x/evm/keeper"

	oraclewasm "github.com/kiichain/kiichain/x/oracle/client/wasm"
	oraclekeeper "github.com/kiichain/kiichain/x/oracle/keeper"

	tokenfactorywasm "github.com/kiichain/kiichain/x/tokenfactory/client/wasm"
	tokenfactorykeeper "github.com/kiichain/kiichain/x/tokenfactory/keeper"
)

func RegisterCustomPlugins(
	epoch *epochkeeper.Keeper,
	tokenfactory *tokenfactorykeeper.Keeper,
	_ *authkeeper.AccountKeeper,
	router wasmkeeper.MessageRouter,
	channelKeeper wasmtypes.ChannelKeeper,
	capabilityKeeper wasmtypes.CapabilityKeeper,
	bankKeeper wasmtypes.Burner,
	unpacker codectypes.AnyUnpacker,
	portSource wasmtypes.ICS20TransferPortSource,
	aclKeeper aclkeeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
	oracleKeeper *oraclekeeper.Keeper,
) []wasmkeeper.Option {
	epochHandler := epochwasm.NewEpochWasmQueryHandler(epoch)
	tokenfactoryHandler := tokenfactorywasm.NewTokenFactoryWasmQueryHandler(tokenfactory)
	evmHandler := evmwasm.NewEVMQueryHandler(evmKeeper)
	oracleHandler := oraclewasm.NewOracleWasmQueryHandler(oracleKeeper)
	wasmQueryPlugin := NewQueryPlugin(epochHandler, tokenfactoryHandler, evmHandler, oracleHandler)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerHandlerOpt := wasmkeeper.WithMessageHandler(
		CustomMessageHandler(router, channelKeeper, capabilityKeeper, bankKeeper, evmKeeper, unpacker, portSource, aclKeeper),
	)

	return []wasm.Option{
		queryPluginOpt,
		messengerHandlerOpt,
	}
}
