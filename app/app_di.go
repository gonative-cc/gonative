package app

import (
	_ "embed" // needed for cosmos-sdk/client/docs/embed.go
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // Import and register pgx driver

	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	_ "cosmossdk.io/indexer/postgres" // register the postgres indexer
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	serverstore "cosmossdk.io/server/v2/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/root"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	lockupdepinject "cosmossdk.io/x/accounts/defaults/lockup/depinject"
	multisigdepinject "cosmossdk.io/x/accounts/defaults/multisig/depinject"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	_ "github.com/cosmos/cosmos-sdk/x/genutil" // genutil effects
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App[T transaction.Tx] struct {
	*runtime.App[T]
	legacyAmino       registry.AminoRegistrar
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry
	store             store.RootStore

	// required keepers during wiring
	// others keepers are all in the app
	UpgradeKeeper *upgradekeeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
}

// Config returns the default app config.
func Config() depinject.Config {
	return depinject.Configs(
		appconfig.Compose(ModuleConfig), // Alternatively use appconfig.LoadYAML(AppConfigYAML)
		runtime.DefaultServiceBindings(),
		codec.DefaultProviders,
		depinject.Provide(
			ProvideRootStoreConfig,
			// inject desired account types:
			multisigdepinject.ProvideAccount,
			basedepinject.ProvideAccount,
			lockupdepinject.ProvideAllLockupAccounts,

			// provide base account options
			basedepinject.ProvideSecp256K1PubKey,
		),
		depinject.Invoke(
			std.RegisterInterfaces,
			std.RegisterLegacyAminoCodec,
		),
	)
}

// NewApp is an App constructor
func NewApp[T transaction.Tx](
	config depinject.Config,
	outputs ...any,
) (*App[T], error) {
	var (
		app          = &App[T]{}
		appBuilder   *runtime.AppBuilder[T]
		storeBuilder root.Builder
		logger       log.Logger

		// merge the AppConfig and other configuration in one config
		appConfig = depinject.Configs(
			Config(),
			config,
			depinject.Supply(
			// ADVANCED CONFIGURATION

			//
			// AUTH
			//
			// For providing a custom function required in auth to generate custom account types
			// add it below. By default the auth module uses simulation.RandomGenesisAccounts.
			//
			// authtypes.RandomGenesisAccountsFn(simulation.RandomGenesisAccounts),
			//
			// For providing a custom a base account type add it below.
			// By default the auth module uses authtypes.ProtoBaseAccount().
			//
			// func() sdk.AccountI { return authtypes.ProtoBaseAccount() },
			//
			// For providing a different address codec, add it below.
			// By default the auth module uses a Bech32 address codec,
			// with the prefix defined in the auth module configuration.
			//
			// func() address.Codec { return <- custom address codec type -> }

			//
			// STAKING
			//
			// For provinding a different validator and consensus address codec, add it below.
			// By default the staking module uses the bech32 prefix provided in the auth config,
			// and appends "valoper" and "valcons" for validator and consensus addresses respectively.
			// When providing a custom address codec in auth, custom address codecs must be provided here as well.
			//
			// func() runtime.ValidatorAddressCodec { return <- custom validator address codec type -> }
			// func() runtime.ConsensusAddressCodec { return <- custom consensus address codec type -> }

			//
			// MINT
			//

			// For providing a custom inflation function for x/mint add here your
			// custom function that implements the minttypes.InflationCalculationFn
			// interface.
			),
			depinject.Provide(
			// if you want to provide a custom public key you
			// can do it from here.
			// Example:
			// 		basedepinject.ProvideCustomPubkey[Ed25519PublicKey]()
			//
			// You can also provide a custom public key with a custom validation function:
			//
			// 		basedepinject.ProvideCustomPubKeyAndValidationFunc(func(pub Ed25519PublicKey) error {
			//			if len(pub.Key) != 64 {
			//				return fmt.Errorf("invalid pub key size")
			//			}
			// 		})
			),
		)
	)

	outputs = append(outputs,
		&logger,
		&storeBuilder,
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.UpgradeKeeper,
		&app.StakingKeeper)

	if err := depinject.Inject(appConfig, outputs...); err != nil {
		return nil, err
	}

	var err error
	app.App, err = appBuilder.Build()
	if err != nil {
		return nil, err
	}

	app.store = storeBuilder.Get()
	if app.store == nil {
		return nil, fmt.Errorf("store builder did not return a db")
	}

	/****  Store Metrics ****/
	/*
		// In order to set store metrics uncomment the below
		storeMetrics, err := metrics.NewMetrics([][]string{{"module", "store"}})
		if err != nil {
			return nil, err
		}
		app.store.SetMetrics(storeMetrics)
	*/
	/****  Module Options ****/

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	app.RegisterUpgradeHandlers()
	if err = app.LoadLatest(); err != nil {
		return nil, err
	}
	return app, nil
}

// AppCodec returns App's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App[T]) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns App's InterfaceRegistry.
func (app *App[T]) InterfaceRegistry() server.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns App's TxConfig.
func (app *App[T]) TxConfig() client.TxConfig {
	return app.txConfig
}

// Store returns the root store.
func (app *App[T]) Store() store.RootStore {
	return app.store
}

// Close overwrites the base Close method to close the stores.
func (app *App[T]) Close() error {
	if err := app.store.Close(); err != nil {
		return err
	}

	return app.App.Close()
}

// ProvideRootStoreConfig deep inject provider for a store
func ProvideRootStoreConfig(config runtime.GlobalConfig) (*root.Config, error) {
	return serverstore.UnmarshalConfig(config)
}
