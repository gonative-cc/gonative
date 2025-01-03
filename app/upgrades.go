package app

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/x/accounts"
	epochstypes "cosmossdk.io/x/epochs/types"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// UpgradeName defines the on-chain upgrade name for the sample App upgrade
// from v0.50.x to v2
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.50.x to v2.
const UpgradeName = "v050-to-v2"

// RegisterUpgradeHandlers for x/gov proposals
func (app *App[T]) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM appmodule.VersionMap) (appmodule.VersionMap, error) {
			return app.ModuleManager().RunMigrations(ctx, fromVM)
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := store.StoreUpgrades{
			Added: []string{
				accounts.ModuleName,
				epochstypes.StoreKey,
				protocolpooltypes.ModuleName,
			},
			Deleted: []string{"crisis"}, // The SDK discontinued the crisis module in v0.52.0
		}

		app.SetStoreLoader(runtime.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
