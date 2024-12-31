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

	// Bech32Prefix account display prefix
	Bech32Prefix = "native"
)

func init() {
	sdk.DefaultBondDenom = BondDenom

	p := Bech32Prefix
	c := sdk.GetConfig()
	c.SetBech32PrefixForAccount(p, sdk.GetBech32PrefixAccPub(p))
	c.SetBech32PrefixForValidator(sdk.GetBech32PrefixValAddr(p), sdk.GetBech32PrefixValPub(p))
	c.SetBech32PrefixForConsensusNode(sdk.GetBech32PrefixConsAddr(p), sdk.GetBech32PrefixConsPub(p))
	c.Seal()
}

// NtivTokenMetadata creates bank Metadata for the NTIV token
func NtivTokenMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Description: "The operational token of the Native blockchain.",
		Base:        BondDenom, // NOTE: must not change
		Name:        "Native",
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
