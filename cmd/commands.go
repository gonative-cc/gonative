package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"cosmossdk.io/client/v2/offchain"
	coreserver "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	runtimev2 "cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	grpcserver "cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/api/grpcgateway"
	"cosmossdk.io/server/v2/api/rest"
	"cosmossdk.io/server/v2/api/telemetry"
	"cosmossdk.io/server/v2/cometbft"
	serverstore "cosmossdk.io/server/v2/store"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdktelemetry "github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2/cli"

	"github.com/gonative-cc/gonative/app"
)

// CommandDependencies is a struct that contains all the dependencies needed to initialize the root command.
// an alternative design could fetch these even later from the command context
type CommandDependencies[T transaction.Tx] struct {
	GlobalConfig  coreserver.ConfigMap
	TxConfig      client.TxConfig
	ModuleManager *runtimev2.MM[T]
	App           *app.App[T]
	// could generally be more generic with serverv2.ServerComponent[T]
	// however, we want to register extra grpc handlers
	ConsensusServer *cometbft.CometBFTServer[T]
	ClientContext   client.Context
}

// InitRootCmd adds sub-commands to the root command.
func InitRootCmd[T transaction.Tx](
	rootCmd *cobra.Command,
	logger log.Logger,
	deps CommandDependencies[T],
) (serverv2.ConfigWriter, error) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(deps.ModuleManager),
		genesisCommand(deps.ModuleManager, deps.App),
		NewTestnetCmd(deps.ModuleManager),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		// add keybase, auxiliary RPC, query, genesis, and tx child commands
		queryCommand(),
		txCommand(),
		keys.Commands(),
		offchain.OffChain(),
		version.NewVersionCommand(),
	)

	// build CLI skeleton for initial config parsing or a client application invocation
	if deps.App == nil {
		if deps.ConsensusServer == nil {
			deps.ConsensusServer = cometbft.NewWithConfigOptions[T](initCometConfig())
		}
		return serverv2.AddCommands[T](
			rootCmd,
			logger,
			io.NopCloser(nil),
			deps.GlobalConfig,
			initServerConfig(),
			deps.ConsensusServer,
			&grpcserver.Server[T]{},
			&serverstore.Server[T]{},
			&telemetry.Server[T]{},
			&rest.Server[T]{},
			&grpcgateway.Server[T]{},
		)
	}

	// store component (not a server)
	storeComponent, err := serverstore.New[T](deps.App.Store(), deps.GlobalConfig)
	if err != nil {
		return nil, err
	}
	restServer, err := rest.New[T](logger, deps.App.App.AppManager, deps.GlobalConfig)
	if err != nil {
		return nil, err
	}

	// consensus component
	if deps.ConsensusServer == nil {
		deps.ConsensusServer, err = cometbft.New(
			logger,
			deps.App.Name(),
			deps.App.Store(),
			deps.App.App.AppManager,
			cometbft.AppCodecs[T]{
				AppCodec:              deps.App.AppCodec(),
				TxCodec:               &client.DefaultTxDecoder[T]{TxConfig: deps.TxConfig},
				LegacyAmino:           deps.ClientContext.LegacyAmino,
				ConsensusAddressCodec: deps.ClientContext.ConsensusAddressCodec,
			},
			deps.App.App.QueryHandlers(),
			deps.App.App.SchemaDecoderResolver(),
			initCometOptions[T](),
			deps.GlobalConfig,
		)
		if err != nil {
			return nil, err
		}
	}

	telemetryServer, err := telemetry.New[T](deps.GlobalConfig, logger, sdktelemetry.EnableTelemetry)
	if err != nil {
		return nil, err
	}

	grpcServer, err := grpcserver.New[T](
		logger,
		deps.App.InterfaceRegistry(),
		deps.App.QueryHandlers(),
		deps.App.Query,
		deps.GlobalConfig,
		grpcserver.WithExtraGRPCHandlers[T](
			deps.ConsensusServer.GRPCServiceRegistrar(
				deps.ClientContext,
				deps.GlobalConfig,
			),
		),
	)
	if err != nil {
		return nil, err
	}

	grpcgatewayServer, err := grpcgateway.New[T](
		logger,
		deps.GlobalConfig,
		deps.App.InterfaceRegistry(),
		deps.App.App.AppManager,
	)
	if err != nil {
		return nil, err
	}
	registerGRPCGatewayRoutes[T](deps, grpcgatewayServer)

	// wire server commands
	return serverv2.AddCommands[T](
		rootCmd,
		logger,
		deps.App,
		deps.GlobalConfig,
		initServerConfig(),
		deps.ConsensusServer,
		grpcServer,
		storeComponent,
		telemetryServer,
		restServer,
		grpcgatewayServer,
	)
}

// genesisCommand builds genesis-related `gonative genesis` command.
func genesisCommand[T transaction.Tx](
	moduleManager *runtimev2.MM[T],
	app *app.App[T],
) *cobra.Command {
	var genTxValidator func([]transaction.Msg) error
	if moduleManager != nil {
		genTxValidator = moduleManager.Modules()[genutiltypes.ModuleName].(genutil.AppModule).GenTxValidator()
	}
	cmd := v2.Commands(
		genTxValidator,
		moduleManager,
		app,
	)

	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.QueryEventForTxCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

// RootCommandPersistentPreRun initializes the root command state
//
//revive:disable:unused-parameter copied from cosmos-sdk
func RootCommandPersistentPreRun(clientCtx client.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// set the default command outputs
		cmd.SetOut(cmd.OutOrStdout())
		cmd.SetErr(cmd.ErrOrStderr())

		clientCtx = clientCtx.WithCmdContext(cmd.Context())
		clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}

		customClientTemplate, customClientConfig := initClientConfig()
		clientCtx, err = config.CreateClientConfig(
			clientCtx, customClientTemplate, customClientConfig)
		if err != nil {
			return err
		}

		return client.SetCmdClientContextHandler(clientCtx, cmd)
	}
}

// registerGRPCGatewayRoutes registers the gRPC gateway routes for all modules and other components
// TODO(@julienrbrt): Eventually, this should removed and directly done within the grpcgateway.Server
// ref: https://github.com/cosmos/cosmos-sdk/pull/22701#pullrequestreview-2470651390
func registerGRPCGatewayRoutes[T transaction.Tx](
	deps CommandDependencies[T],
	server *grpcgateway.Server[T],
) {
	// those are the extra services that the CometBFT server implements (server/v2/cometbft/grpc.go)
	cmtservice.RegisterGRPCGatewayRoutes(deps.ClientContext, server.GRPCGatewayRouter)
	_ = nodeservice.RegisterServiceHandlerClient(context.Background(), server.GRPCGatewayRouter, nodeservice.NewServiceClient(deps.ClientContext))
	_ = txtypes.RegisterServiceHandlerClient(context.Background(), server.GRPCGatewayRouter, txtypes.NewServiceClient(deps.ClientContext))
}
