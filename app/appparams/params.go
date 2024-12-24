package appparams

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Name defines the application name.
	Name = "gonative"

	// BondDenom defines the native staking token denomination.
	// NOTE: it is used by IBC, and must not change to avoid token migration in all IBC chains.
	BondDenom = "untiv"
)

func init() {
	sdk.DefaultBondDenom = BondDenom
}
