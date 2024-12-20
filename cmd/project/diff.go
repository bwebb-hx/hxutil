package projectCmd

import (
	"github.com/bwebb-hx/hxutil/internal/project"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff two projects in Hexabase",
	Long:  `Diff two projects in Hexabase.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.PrintErrln("2 project IDs required")
			return
		}
		pid1 := args[0]
		pid2 := args[1]
		project.Diff(pid1, pid2)
	},
}

func init() {
	Cmd.AddCommand(diffCmd)
}
