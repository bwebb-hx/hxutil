package cmd

import (
	"os"

	actionCmd "github.com/bwebb-hx/hxutil/cmd/action"
	apiCmd "github.com/bwebb-hx/hxutil/cmd/api"
	projectCmd "github.com/bwebb-hx/hxutil/cmd/project"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "hxutil",
	Short: "a collection of utility tools for hexabase!",
	Long:  `A collection of utility tools for hexabase! Includes tools to test APIs, manage ActionScripts for projects, and more things to come.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(actionCmd.Cmd)
	RootCmd.AddCommand(apiCmd.Cmd)
	RootCmd.AddCommand(projectCmd.Cmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hxutil.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
