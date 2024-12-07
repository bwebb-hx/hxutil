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
	Short: "Diff ActionScripts between a local repository and the remote version saved in Hexabase.",
	Long: `Diff ActionScripts between a local repository and the remote version saved in Hexabase.
This command is useful for ensuring that ActionScripts that are saved in Hexabase are properly tracked in your version control.
It is expected that ActionScripts are saved in your project using the display ID of the action, suffixed with either "pre" or "post" depending on the script type.
This command will recursively search all directories under the directory it is called in.`,
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
	diffCmd.Flags().StringVarP(&dir, "dir", "d", ".", "path to a project directory to diff. defaults to the current directory.")
	Cmd.AddCommand(diffCmd)
}
