package cmd_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/core/transaction"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/require"
	
	"github.com/gonative-cc/gonative/cmd"
)

func TestInitTestFilesCmd(t *testing.T) {
	args := []string{
		"testnet", // Test the testnet init-files command
		"init-files",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest), // Set keyring-backend to test
	}
	rootCmd, err := cmd.NewRootCmd[transaction.Tx](args...)
	require.NoError(t, err)
	rootCmd.SetArgs(args)
	require.NoError(t, rootCmd.Execute())
}
