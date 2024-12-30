package overwrite

import (
	"encoding/json"

	"github.com/gonative-cc/gonative/app/appparams"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"github.com/cosmos/cosmos-sdk/codec"

	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
)

// RegisterModules overwrites registered modules
func RegisterModules() {
	appconfig.RegisterModule(
		&bankmodulev1.Module{},
		appconfig.Provide(ProvideBankModule),
		appconfig.Invoke(bank.InvokeSetSendRestrictions),
	)
}

// BankModuleOutputs for depinject provider
type BankModuleOutputs struct {
	depinject.Out

	BankKeeper bankkeeper.BaseKeeper
	Module     appmodule.AppModule
}

// ProvideBankModule is a depinject provider for the overwritten bank module.
func ProvideBankModule(in bank.ModuleInputs) BankModuleOutputs {
	pm := bank.ProvideModule(in)
	m := NewBankAppModule(in.Cdc, pm.BankKeeper, in.AccountKeeper)

	return BankModuleOutputs{
		BankKeeper: pm.BankKeeper,
		Module:     m,
	}
}

// BankAppModule wraps SDK bank AppModule.
type BankAppModule struct {
	bank.AppModule

	cdc codec.Codec
}

// NewBankAppModule creates a new AppModule object
func NewBankAppModule(cdc codec.Codec, keeper bankkeeper.Keeper, accountKeeper banktypes.AccountKeeper) BankAppModule {
	am := bank.NewAppModule(cdc, keeper, accountKeeper)
	return BankAppModule{
		AppModule: am,
		cdc:       cdc,
	}
}

// DefaultGenesis returns default genesis state as raw bytes for the bank module.
func (am BankAppModule) DefaultGenesis() json.RawMessage {
	g := banktypes.DefaultGenesisState()
	g.DenomMetadata = append(g.DenomMetadata, appparams.NtivTokenMetadata())
	return am.cdc.MustMarshalJSON(g)
}
