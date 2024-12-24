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
		pid1, pid2 := "", ""
		if len(args) > 0 {
			pid1 = args[0]
		}
		if len(args) > 1 {
			pid2 = args[1]
		}
		project.Diff(pid1, pid2)
	},
}

func init() {
	Cmd.AddCommand(diffCmd)
}
