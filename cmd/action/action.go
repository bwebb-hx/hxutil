package actionCmd

import (
	"github.com/spf13/cobra"
)

// Root for the action command group
var Cmd = &cobra.Command{
	Use:   "action",
	Short: "Utilities for actions and actionscripts",
	Long: `Utilities for actions and actionscripts.

Commands:

- diff: check for differences in the action scripts for a project between local and remote.`,
	// Uncomment the following line if the bare command
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
}
