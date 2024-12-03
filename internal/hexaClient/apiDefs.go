package hexaclient

type ApiEndpoint struct {
	URI            string
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
	Method:         "POST",
	RequireToken:   false,
	RequirePayload: true,
}

var GetWorkspacesAPI = ApiEndpoint{
	URI:            "/api/v0/workspaces",
	Method:         "GET",
	RequireToken:   true,
	RequirePayload: false,
}

// https://apidoc.hexabase.com/en/docs/v0/datastores/GetActions
var GetActionsAPI = ApiEndpoint{
	URI:            "/api/v0/datastores/%s/actions",
	Method:         "GET",
	RequireToken:   true,
	RequirePayload: false,
}
