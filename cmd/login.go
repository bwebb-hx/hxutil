/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"

	hexaclient "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login, and get an auth token if successful, for given credentials for Hexabase.",
	Long: `Login, and get an auth token if successful, for given credentials for Hexabase.
	Useful for confirming login credentials, or for obtaining a token for test purposes.
	
	Usage Examples:
	
	# use credentials set in config
	hxutil login
	
	# specify email and password
	hxutil login "someUser@hexabase.com" "password123"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("must specify email; config usage not yet supported")
		}

		email := args[0]
		password := args[1]

		fmt.Println("email:", email)
		fmt.Println("password:", password)

		token := hexaclient.Login(email, password)
		if token == "" {
			return errors.New("login failed; no token generated.")
		}

		fmt.Println("Login successful! Token:")
		fmt.Println(token)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
