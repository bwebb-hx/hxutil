package apiCmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	hexaclient "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/spf13/cobra"
)

var (
	method string
	body   string
	auth   bool
)

var callCmd = &cobra.Command{
	Use:   "call",
	Short: "Calls a given URI as a one-off test, and shows the response",
	Long: `Calls a given URI as a one-off test, and shows the response.

You can optionally enter variable naems in the URI, such as ":p-id" (or :project-id), :d-id (:datastore-id), etc
which will be automatically replaced with the config variables, user, etc.

# do a GET request, with config variables (:d-id) applied to the URI and the auth flag set
hxutil api call /api/v0/datastores/:d-id/actions -a

# do a POST call with a payload, without authorization
hxutil api call /api/v0/login -m POST -b '{ "email": "user@company.com", "password": "xyz" }'`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.PrintErrln("URI is required")
			return
		}
		uri := args[0]

		token := ""
		if auth {
			token = hexaclient.Login(hexaclient.TestAccUser, hexaclient.TestAccPass)
		}

		if method == "GET" {
			if body != "" {
				fmt.Println("Warning: given body not used since this is a GET request. Use the --method flag to make a POST request.")
			}
			resp, err := hexaclient.GetApi(uri, nil, token)
			if err != nil {
				cmd.PrintErrln("Error occurred in API execution:", err)
				return
			}
			formatResponse(resp)
			return
		} else if method == "POST" {
			resp, err := hexaclient.PostApi(uri, []byte(body), token)
			if err != nil {
				cmd.PrintErrln("Error occurred in APIP execution:", err)
				return
			}
			formatResponse(resp)
			return
		}
		// unsupported method
		cmd.PrintErrln("Provided method not recognized or supported:", method)
	},
}

func init() {
	callCmd.Flags().StringVarP(&method, "method", "m", "GET", "method to use when calling the API.")
	callCmd.Flags().StringVarP(&body, "body", "b", "", "body payload to pass when calling the API. only used for POST requests.")
	callCmd.Flags().BoolVarP(&auth, "auth", "a", false, "if flag is set, config is used to get hexabase auth token to pass in authorization header.")

	Cmd.AddCommand(callCmd)
}

func formatResponse(resp []byte) {
	// check if the response is a json
	rawString := strings.TrimSpace(string(resp))
	if rawString[0] == '{' && rawString[len(rawString)-1] == '}' {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(resp, &jsonData); err != nil {
			log.Fatal("failed to unmarshal json:", err)
		}
		// great, we've successfully unmarshalled the data. let's remarshal it with some indentation now to make it readable
		indentJsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			log.Fatal("failed to remarshal json:", err)
		}
		fmt.Println("Response (formatted JSON):")
		fmt.Println(string(indentJsonBytes))
		return
	}
	// response doesn't appear to be a (single) json. so just show the raw string data
	fmt.Println("Response (raw string):")
	fmt.Println(rawString)
}
