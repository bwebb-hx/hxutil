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
