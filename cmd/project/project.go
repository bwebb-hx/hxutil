package projectCmd

import "github.com/spf13/cobra"

// Root for the action command group
var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Utilities for Hexabase projects",
	Long:  `Utilities for Hexabase projects.`,
	// Uncomment the following line if the bare command
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
}
