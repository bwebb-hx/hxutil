package hexaclient

// (POST)
const LoginURI = "/api/v0/login"

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Action APIs

// (GET)
//
// Insert: database ID
//
// Docs:
// https://apidoc.hexabase.com/en/docs/v0/datastores/GetActions
const GetActionsURI = "/api/v0/datastores/%s/actions"
