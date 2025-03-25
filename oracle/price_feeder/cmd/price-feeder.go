package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	input "github.com/cosmos/cosmos-sdk/client/input"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/kiichain/kiichain/oracle/price_feeder/config"
	"github.com/kiichain/kiichain/oracle/price_feeder/oracle"
	"github.com/kiichain/kiichain/oracle/price_feeder/oracle/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	LOG_LEVEL_JSON = "json"
	LOG_LEVEL_TEXT = "text"

	FLAG_LOG_LEVEL  = "log-level"
	FLAG_LOG_FORMAT = "log-format"

	envVariablePass = "PRICE_FEEDER_PASS"
)

var rootCmd = &cobra.Command{
	Use:   "price-feeder [config-file]",
	Args:  cobra.ExactArgs(1),
	Short: "price-feeder is a process which provides prices data to the oracle module",
	Long: `price-feeder is a process that validators must run in order to provide oracle with 
price information. The price-feeder obtains price information from various reliable data 
sources, e.g. exchanges, then, submits vote messages following the oracle voting procedure.`,
	RunE: priceFeederCmdHandler,
}

// init is executed automatically when by the Golang work flow and adds the version subcommand
// and persistent flags
func init() {
	rootCmd.PersistentFlags().String(FLAG_LOG_LEVEL, zerolog.InfoLevel.String(), "logging level")
	rootCmd.PersistentFlags().String(FLAG_LOG_FORMAT, LOG_LEVEL_TEXT, "logging format; must be either json or text")

	// add subcomands
	rootCmd.AddCommand(CmdgetVersion())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// priceFeederCmdHandler init the price feeder
func priceFeederCmdHandler(cmd *cobra.Command, args []string) error {
	// get value from the log level cmd flag
	logLvlStr, err := cmd.Flags().GetString(FLAG_LOG_LEVEL)
	if err != nil {
		return err
	}

	// get value from the log format cmd flag
	logFormatStr, err := cmd.Flags().GetString(FLAG_LOG_FORMAT)
	if err != nil {
		return err
	}

	logLvl, err := zerolog.ParseLevel(logLvlStr)
	if err != nil {
		return err
	}

	// set the log format based on the flags
	var logWriter io.Writer
	switch strings.ToLower(logFormatStr) {

	case LOG_LEVEL_JSON:
		logWriter = os.Stderr

	case LOG_LEVEL_TEXT:
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}

	default:
		return fmt.Errorf("invalid logging format: %s", logFormatStr)
	}

	// create looger
	logger := zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()

	// pase configurations from the config file to Config struct
	cfg, err := config.ParseConfig(args[0])
	if err != nil {
		return err
	}

	// Set prefixes (cosmos will analize accounts with these prefixes)
	accountPubKeyPrefix := cfg.Account.Prefix + "pub"
	validatorAddressPrefix := cfg.Account.Prefix + "valoper"
	validatorPubKeyPrefix := cfg.Account.Prefix + "valoperpub"
	consNodeAddressPrefix := cfg.Account.Prefix + "valcons"
	consNodePubKeyPrefix := cfg.Account.Prefix + "valconspub"

	// Set and seal config
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(cfg.Account.Prefix, accountPubKeyPrefix)           // accounts must have this prefix
	sdkConfig.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)   // validators' address must have this prefix
	sdkConfig.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix) // node address must have this prefix
	sdkConfig.Seal()                                                                       // the previous configuration cannont be changed further

	// Create context and goroutines group
	ctx, cancel := context.WithCancel(cmd.Context())
	group, ctx := errgroup.WithContext(ctx)

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(cancel, logger)

	// get rpc timeout from config
	rpcTimeout, err := time.ParseDuration(cfg.RPC.RPCTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse RPC timeout: %w", err)
	}

	// Gather password via env variable or std input
	keyringPass, err := getKeyringPassword()
	if err != nil {
		return err
	}

	// Retry creating oracle client for 5 seconds
	var oracleClient client.OracleClient
	for i := 0; i < 5; i++ {
		oracleClient, err = client.NewOracleClient(
			ctx,
			logger,
			cfg.Account.ChainID,
			cfg.Keyring.Backend,
			cfg.Keyring.Dir,
			keyringPass,
			cfg.RPC.TMRPCEndpoint,
			rpcTimeout,
			cfg.Account.Address,
			cfg.Account.Validator,
			cfg.Account.FeeGranter,
			cfg.RPC.GRPCEndpoint,
			cfg.GasAdjustment,
			cfg.GasPrices,
		)
		if err != nil {
			// sleep for a second before retrying
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return fmt.Errorf("error creating oracle client: %w", err)
	}

	// get provider timeout from config
	providerTimeout, err := time.ParseDuration(cfg.ProviderTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse provider timeout: %w", err)
	}

	// create a map with the deviation by denom from config file
	deviations := make(map[string]sdk.Dec, len(cfg.Deviations))
	for _, deviation := range cfg.Deviations {
		threshold, err := sdk.NewDecFromStr(deviation.Threshold)
		if err != nil {
			return err
		}
		deviations[deviation.Base] = threshold
	}

	// create a map with the endpoitns listed on the config file
	endpoints := make(map[string]config.ProviderEndpoint, len(cfg.ProviderEndpoints))
	for _, endpoint := range cfg.ProviderEndpoints {
		endpoints[endpoint.Name] = endpoint
	}

	// create new oracle instance
	oracle := oracle.New(
		logger,
		oracleClient,
		cfg.CurrencyPairs,
		providerTimeout,
		deviations,
		endpoints,
		cfg.Healthchecks,
	)

	// start the process that calculates oracle prices and votes
	group.Go(func() error {
		return startPriceOracle(ctx, logger, oracle)
	})

	// Block main process until all spawned goroutines have gracefully exited and
	// signal has been captured in the main process or if an error occurs.
	return group.Wait()
}

// getKeyringPassword obtains the keyring password from the env var or stdin
func getKeyringPassword() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	pass := os.Getenv(envVariablePass)
	if pass == "" {
		return input.GetString("Enter keyring password", reader)
	}
	return pass, nil
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal(cancel context.CancelFunc, logger zerolog.Logger) {
	// create channel to store the signals
	sigCh := make(chan os.Signal, 1)

	// stay alert for SIGTERM and SIGINT signals from the OS and store on the channel
	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	// launch a goroutine to handle the signal reception
	go func() {
		sig := <-sigCh // wait until the channel return a value
		logger.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		cancel() // execute cancel and cancel the main process
	}()
}

// startPriceOracle initialize a goroutine with the price-feeder
func startPriceOracle(ctx context.Context, logger zerolog.Logger, oracle *oracle.Oracle) error {
	// channel to receive errors from the price-feeder
	srvErrCh := make(chan error, 1)

	// launch price-feeder as goroutine
	go func() {
		logger.Info().Msg("starting price-feeder oracle...")
		srvErrCh <- oracle.Start(ctx)
	}()

	// stay tuned for errors on the context or price feeder
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("shutting down price-feeder oracle...")
			return nil

		case err := <-srvErrCh:
			logger.Err(err).Msg("error starting the price-feeder oracle")
			oracle.Stop()
			return err
		}
	}
}
