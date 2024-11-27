package apiCmd

import (
	hexaclient "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the Hexabase APIs",
	Long: `Run tests on the Hexabase APIs to see how they are currently performing.

Usage:
hxutil api test`,
	Run: func(cmd *cobra.Command, args []string) {
		hexaclient.RunStatusCheck()
	},
}

func init() {
	Cmd.AddCommand(testCmd)
}
