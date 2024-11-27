package apiCmd

import "github.com/spf13/cobra"

// Root for the action command group
var Cmd = &cobra.Command{
	Use:   "api",
	Short: "Utilities for testing Hexabase APIs",
	Long: `Utilities for testing Hexabase APIs.

Commands:

- test: run tests to see how APIs are currently performing.`,
	// Uncomment the following line if the bare command
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
}
