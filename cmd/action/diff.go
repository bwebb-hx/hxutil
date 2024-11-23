package actionCmd

import (
	"fmt"
	"path/filepath"

	"github.com/bwebb-hx/hxutil/internal/action"
	"github.com/spf13/cobra"
)

var dir string

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := dir

		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Printf("Error resolving path: %d\n", err)
			return
		}

		action.DiffActionScripts(absPath)
	},
}

func init() {
	diffCmd.Flags().StringVarP(&dir, "dir", "d", ".", "path to a project directory to diff")
	Cmd.AddCommand(diffCmd)
}
