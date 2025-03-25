package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/rs/zerolog"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

type (
	// OracleClient defines a structure that interact with the kiichain node.
	OracleClient struct {
		Logger              zerolog.Logger
		ChainID             string
		KeyringBackend      string
		KeyringDir          string
		KeyringPass         string
		TMRPC               string
		RPCTimeout          time.Duration
		OracleAddr          sdk.AccAddress
		OracleAddrString    string
		ValidatorAddr       sdk.ValAddress
		ValidatorAddrString string
		FeeGranterAddr      sdk.AccAddress
		Encoding            simappparams.EncodingConfig
		GasPrices           string
		GasAdjustment       float64
		GRPCEndpoint        string
		KeyringPassphrase   string
		BlockHeightEvents   chan int64

		// MockBroadcastTx allows for a basic mock without refactoring this to an interface
		MockBroadcastTx func(clientCtx client.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error)
	}

	passReader struct {
		pass string
		buf  *bytes.Buffer
	}
)

// NewOracleClient creates a new instance of the OracleClient
func NewOracleClient(
	ctx context.Context,
	logger zerolog.Logger,
	chainID string,
	keyringBackend string,
	keyringDir string,
	keyringPass string,
	tmRPC string,
	rpcTimeout time.Duration,
	oracleAddrString string,
	validatorAddrString string,
	feeGranterAddrString string,
	grpcEndpoint string,
	gasAdjustment float64,
	gasPrices string,
) (OracleClient, error) {
	// get the account which performs the transaction
	oracleAddr, err := sdk.AccAddressFromBech32(oracleAddrString)
	if err != nil {
		return OracleClient{}, err
	}

	// get the account who will pay the gas
	feegrantAddr, _ := sdk.AccAddressFromBech32(feeGranterAddrString)

	// create client
	oracleClient := OracleClient{
		Logger:              logger.With().Str("module", "oracle_client").Logger(),
		ChainID:             chainID,
		KeyringBackend:      keyringBackend,
		KeyringDir:          keyringDir,
		KeyringPass:         keyringPass,
		TMRPC:               tmRPC, // tendermint endpoint
		RPCTimeout:          rpcTimeout,
		OracleAddr:          oracleAddr,
		OracleAddrString:    oracleAddrString,
		ValidatorAddr:       sdk.ValAddress(validatorAddrString),
		ValidatorAddrString: validatorAddrString,
		FeeGranterAddr:      feegrantAddr,
		Encoding:            simapp.MakeTestEncodingConfig(),
		GasAdjustment:       gasAdjustment,
		GRPCEndpoint:        grpcEndpoint,
		GasPrices:           gasPrices,
		BlockHeightEvents:   make(chan int64, 1),
	}

	// creates the cosmos client context based on the oracle client
	clientCtx, err := oracleClient.CreateClientContext()
	if err != nil {
		return OracleClient{}, err
	}

	// get block height from the rpc connection
	blockHeight, err := rpc.GetChainHeight(clientCtx)
	if err != nil {
		return OracleClient{}, err
	}

	// create a chain tracker (HeightUpdater is used to subscribe to event of new block generated)
	chainHeightUpdater := HeightUpdater{
		Logger:        logger,
		LastHeight:    blockHeight,
		ChBlockHeight: oracleClient.BlockHeightEvents,
	}

	// start tracking the chain for new block events and update the height
	err = chainHeightUpdater.Start(ctx, clientCtx.Client, oracleClient.Logger)
	if err != nil {
		return OracleClient{}, err
	}

	return oracleClient, nil
}

// newPassReader returns a reader obj with the password from env
func newPassReader(pass string) io.Reader {
	return &passReader{
		pass: pass,
		buf:  new(bytes.Buffer),
	}
}

func (r *passReader) Read(p []byte) (n int, err error) {
	n, err = r.buf.Read(p)
	if err == io.EOF || n == 0 {
		r.buf.WriteString(r.pass + "\n")

		n, err = r.buf.Read(p)
	}

	return n, err
}

// BroadcastTx attempts to broadcast a signed transaction in best effort mode.
// Retry is not needed since we are doing this for every new block as fast as we could.
// Ref: https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L230
//
// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
//
// It will return an error upon failure. We maintain a local account sequence number in txAccount
// and we manually increment the sequence number by 1 if the previous broadcastTx succeed.
func (oc OracleClient) BroadcastTx(
	clientCtx client.Context,
	msgs ...sdk.Msg) (*sdk.TxResponse, error) {

	// this allows for basic mocking without refactoring this to an interface (much larger change)
	if oc.MockBroadcastTx != nil {
		return oc.MockBroadcastTx(clientCtx, msgs...)
	}

	// get current time (for measuring the time taken to send the transaction)
	startTime := time.Now()
	defer telemetry.MeasureSince(startTime, "latency", "broadcast")

	// create transaction factory
	txf, err := oc.CreateTxFactory()
	if err != nil {
		return nil, err
	}

	// get account number and next sequence
	txAccountInfo := NewAccountInfo()
	txf, err = txAccountInfo.ObtainAccountInfo(clientCtx, txf, oc.Logger)
	if err != nil {
		return nil, err
	}

	// Build unsigned tx
	transaction, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	// Sign the transaction
	err = tx.Sign(txf, clientCtx.GetFromName(), transaction, true)
	if err != nil {
		return nil, err
	}

	// convert transaction to bytes to be sent
	txBytes, err := clientCtx.TxConfig.TxEncoder()(transaction.GetTx())
	if err != nil {
		return nil, err
	}

	oc.Logger.Info().Msg(fmt.Sprintf("Sending broadcastTx with account sequence number %d", txf.Sequence()))

	// broadcast transaction
	resp, err := clientCtx.BroadcastTx(txBytes)
	if resp != nil && resp.Code != 0 && resp.Code != sdkerrors.ErrAlreadyExists.ABCICode() {
		err = fmt.Errorf("received error response code %d from broadcast tx: %s", resp.Code, resp.Logs.String())
		return resp, err
	}
	if err != nil {
		// When error happen, it could be that the sequence number are mismatching
		// We need to reset sequence number to query the latest value from the chain
		txAccountInfo.ShouldResetSequence = true
		return resp, err
	}
	// Only increment sequence number if we successfully broadcast the previous transaction
	txAccountInfo.AccountSequence++
	return resp, err

}

// CreateClientContext creates an SDK client Context instance used for transaction
// generation, signing and broadcasting.
func (oc OracleClient) CreateClientContext() (client.Context, error) {
	// get keyring password from selected input
	var keyringInput io.Reader
	if len(oc.KeyringPass) > 0 {
		keyringInput = newPassReader(oc.KeyringPass)
	} else {
		keyringInput = os.Stdin
	}

	// create a new keyring
	kr, err := keyring.New("kiichain3", oc.KeyringBackend, oc.KeyringDir, keyringInput)
	if err != nil {
		return client.Context{}, err
	}

	// create a tendermint HTTP client
	httpClient, err := tmjsonclient.DefaultHTTPClient(oc.TMRPC)
	if err != nil {
		return client.Context{}, err
	}

	httpClient.Timeout = oc.RPCTimeout

	// create a tendermint RPC client
	tmRPC, err := rpchttp.NewWithClient(oc.TMRPC, httpClient)
	if err != nil {
		return client.Context{}, err
	}

	// get keyring from the info from the oracle addr
	keyInfo, err := kr.KeyByAddress(oc.OracleAddr)
	if err != nil {
		return client.Context{}, err
	}

	// create a cosmos client context
	clientCtx := client.Context{
		ChainID:           oc.ChainID,
		JSONCodec:         oc.Encoding.Marshaler,
		InterfaceRegistry: oc.Encoding.InterfaceRegistry,
		Output:            os.Stderr,
		BroadcastMode:     flags.BroadcastSync,
		TxConfig:          oc.Encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		Codec:             oc.Encoding.Marshaler,
		LegacyAmino:       oc.Encoding.Amino,
		Input:             os.Stdin,
		NodeURI:           oc.TMRPC,
		Client:            tmRPC,
		Keyring:           kr,
		FromAddress:       oc.OracleAddr,
		FromName:          keyInfo.GetName(),
		From:              keyInfo.GetName(),
		OutputFormat:      "json",
		UseLedger:         false,
		Simulate:          false,
		GenerateOnly:      false,
		Offline:           false,
		SkipConfirm:       true,
		FeeGranter:        oc.FeeGranterAddr,
	}

	return clientCtx, nil
}

// CreateTxFactory creates an SDK Factory instance used for transaction
// generation, signing and broadcasting.
func (oc OracleClient) CreateTxFactory() (tx.Factory, error) {
	// create cosmos context
	clientCtx, err := oc.CreateClientContext()
	if err != nil {
		return tx.Factory{}, err
	}

	// craete a transaction
	txFactory := tx.Factory{}.
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(oc.ChainID).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(oc.GasAdjustment).
		WithGasPrices(oc.GasPrices).
		WithKeybase(clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)

	return txFactory, nil
}
