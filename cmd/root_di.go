package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"

	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"

	"github.com/gonative-cc/gonative/app"
)

// NewRootCmd creates a root command
//
//revive:disable:cyclomatic mostly copied from cosmos-sdk
func NewRootCmd[T transaction.Tx](
	args ...string,
) (*cobra.Command, error) {
	rootCommand := &cobra.Command{
		Use:           "gonative",
		Short:         "Native Node Application",
		SilenceErrors: true,
	}
	configWriter, err := InitRootCmd(rootCommand, log.NewNopLogger(), CommandDependencies[T]{})
	if err != nil {
		return nil, err
	}
	factory, err := serverv2.NewCommandFactory(
		serverv2.WithConfigWriter(configWriter),
		serverv2.WithStdDefaultHomeDir(".gonative"),
		serverv2.WithLoggerFactory(serverv2.NewLogger),
	)
	if err != nil {
		return nil, err
	}

	nodeCmds := nodeservice.NewNodeCommands()
	autoCLIModuleOpts := make(map[string]*autocliv1.ModuleOptions)
	autoCLIModuleOpts[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()
	autoCliOpts, err := autocli.NewAppOptionsFromConfig(
		depinject.Configs(app.Config(), depinject.Supply(runtime.GlobalConfig{})),
		autoCLIModuleOpts,
	)
	if err != nil {
		return nil, err
	}

	if err = autoCliOpts.EnhanceRootCommand(rootCommand); err != nil {
		return nil, err
	}
	subCommand, configMap, logger, err := factory.ParseCommand(rootCommand, args)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return rootCommand, nil
		}
		return nil, err
	}

	var (
		moduleManager   *runtime.MM[T]
		clientCtx       client.Context
		a               *app.App[T]
		depinjectConfig = depinject.Configs(
			depinject.Supply(logger, runtime.GlobalConfig(configMap)),
			depinject.Provide(ProvideClientContext),
		)
	)
	if serverv2.IsAppRequired(subCommand) {
		// server construction
		a, err = app.NewApp[T](depinjectConfig, &autoCliOpts, &moduleManager, &clientCtx)
		if err != nil {
			return nil, err
		}
	} else {
		// client construction
		if err = depinject.Inject(
			depinject.Configs(
				app.Config(),
				depinjectConfig,
			),
			&autoCliOpts, &moduleManager, &clientCtx,
		); err != nil {
			return nil, err
		}
	}

	commandDeps := CommandDependencies[T]{
		GlobalConfig:  configMap,
		TxConfig:      clientCtx.TxConfig,
		ModuleManager: moduleManager,
		App:           a,
		ClientContext: clientCtx,
	}
	rootCommand = &cobra.Command{
		Use:               rootCommand.Use,
		Short:             rootCommand.Short,
		SilenceErrors:     true,
		PersistentPreRunE: RootCommandPersistentPreRun(clientCtx),
	}
	factory.EnhanceRootCommand(rootCommand)
	_, err = InitRootCmd(rootCommand, logger, commandDeps)
	if err != nil {
		return nil, err
	}
	autoCliOpts.ModuleOptions = autoCLIModuleOpts
	if err := autoCliOpts.EnhanceRootCommand(rootCommand); err != nil {
		return nil, err
	}

	return rootCommand, nil
}
