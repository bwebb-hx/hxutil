package hexaclient

type ApiEndpoint struct {
	URI            string
	DisplayURI     string // URI to show as a general representation
	Method         string
	RequireToken   bool
	RequirePayload bool
}

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var LoginAPI = ApiEndpoint{
	URI:            "/api/v0/login",
	DisplayURI:     "/api/v0/login",
	Method:         POST,
	RequireToken:   false,
	RequirePayload: true,
}

var GetWorkspacesAPI = ApiEndpoint{
	URI:            "/api/v0/workspaces",
	DisplayURI:     "/api/v0/workspaces",
	Method:         GET,
	RequireToken:   true,
	RequirePayload: false,
}

// https://apidoc.hexabase.com/en/docs/v0/datastores/GetActions
var GetActionsAPI = ApiEndpoint{
	URI:            "/api/v0/datastores/%s/actions",
	DisplayURI:     "/api/v0/datastores/:d_id/actions",
	Method:         GET,
	RequireToken:   true,
	RequirePayload: false,
}

type GetActionsResponse []struct {
	W_ID           string `json:"workspace_id"`
	P_ID           string `json:"project_id"`
	D_ID           string `json:"datastore_id"`
	ActionID       string `json:"action_id"`
	IsStatusAction bool   `json:"is_status_action"`
	DisplayID      string `json:"display_id"`
	Name           string `json:"name"`
}

// Based on: https://github.com/hexabase/hexabase-cli/blob/master/src/commands/actions/scripts/download.ts
//
// Query Params:
//
// - script_type: "pre" or "post"
var DownloadActionScriptAPI = ApiEndpoint{
	URI:            "/api/v0/actions/%s/actionscripts/download",
	DisplayURI:     "/api/v0/actions/:action_id/actionscripts/download",
	Method:         GET,
	RequireToken:   true,
	RequirePayload: false,
}

var GetApplicationScriptVariableAPI = ApiEndpoint{
	URI:            "/api/v0/applications/%s/script/%s",
	DisplayURI:     "/api/v0/applications/:app-id/script/:var-name",
	Method:         GET,
	RequireToken:   true,
	RequirePayload: false,
}

var GetDatastoresAPI = ApiEndpoint{
	URI:            "/api/v0/applications/%s/datastores",
	DisplayURI:     "/api/v0/applications/:project-id/datastores",
	Method:         GET,
	RequireToken:   true,
	RequirePayload: false,
}

type GetDatastoresResponse []struct {
	DatastoreID string `json:"datastore_id"`
	Name        string `json:"name"`
	DisplayID   string `json:"display_id"`
	Deleted     bool   `json:"deleted"`
	Imported    bool   `json:"imported"`
	Uploading   bool   `json:"uploading"`
}

// APP.HEXABASE.COM APIS
// The following are not officially published APIs, but ones that I've found while investigating the
// hexabase management console site using the network inspector
// I will prefix all of these APIs with "UN" ("unofficial") until they are replaced with officially documented APIs.

// (UNOFFICIAL)
//
// Query Params:
//
// - p_id: application (project) ID where the function is defined
var UN_GetFunctionActionScriptAPI = ApiEndpoint{
	URI:            "https://app.hexabase.com/v1/api/get_action_scripts",
	DisplayURI:     "(UN) /v1/api/get_action_scripts",
	Method:         GET,
	RequireToken:   true,
	RequirePayload: false,
}

type UN_GetFunctionScripScriptResponse []struct {
	ID         string `json:"_id"`
	AID        string `json:"a_id"`
	FunctionID string `json:"fn_id"`
	DID        string `json:"d_id"`
	PID        string `json:"p_id"`
	WID        string `json:"w_id"`
	Pre        struct {
		Script     string `json:"script"`
		TimeoutSec int    `json:"timeout_sec"`
	} `json:"pre"`
	Name         string `json:"name"`
	DisplayID    string `json:"display_id"`
	WaitResponse bool   `json:"wait_response"`
}
