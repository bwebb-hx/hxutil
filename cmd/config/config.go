package configCmd

import (
	"fmt"

	"github.com/bwebb-hx/hxutil/internal/config"
	"github.com/bwebb-hx/hxutil/internal/utils"
	"github.com/spf13/cobra"
)

// Root for the action command group
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage hxutil configuration",
	Long:  `Manage hxutil configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		// starts an interface to let users manage config
		fmt.Println("HXUTIL Config Console: Coming soon!")
		path := config.ConfigFilePath()
		utils.Info("config path: "+path, "Edit this file to make changes to configuration")
	},
}

func init() {
}
