package appparams

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "cosmossdk.io/x/bank/types"
)

const (
	// AppName defines the application name.
	AppName = "gonative"

	// BondDenom defines the native staking token denomination.
	// NOTE: it is used by IBC, and must not change to avoid token migration in all IBC chains.
	BondDenom = "untiv"
	// DisplayDenom defines the name, symbol, and display value of the NTIV token.
	DisplayDenom = "NTIV"
)

func init() {
	sdk.DefaultBondDenom = BondDenom
}

// NtivTokenMetadata creates bank Metadata for the NTIV token
func NtivTokenMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Description: "The operational token of the Native blockchain.",
		Base:        BondDenom, // NOTE: must not change
		Name:        DisplayDenom,
		Display:     DisplayDenom,
		Symbol:      DisplayDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BondDenom,
				Exponent: 0,
				Aliases:  []string{"microntiv"},
			}, {
				Denom:    DisplayDenom,
				Exponent: 6,
				Aliases:  []string{},
			},
		},
	}
}
