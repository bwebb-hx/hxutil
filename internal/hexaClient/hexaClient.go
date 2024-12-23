package hexaclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwebb-hx/hxutil/internal/utils"
)

var baseURL = "https://api.hexabase.com"

var Token string = ""

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

func SetBaseUrl(url string) {
	baseURL = url
}

func PostApi(uri string, body []byte) ([]byte, error) {
	if !strings.Contains(uri, "http") {
		uri = fmt.Sprintf("%s%s", baseURL, uri)
	}

	req, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Token))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func GetApi(uri string, queryParams map[string]string) ([]byte, error) {
	if !strings.Contains(uri, "http") {
		uri = fmt.Sprintf("%s%s", baseURL, uri)
	}

	if queryParams != nil {
		params := make([]string, 0)
		uri += "?"
		for param, value := range queryParams {
			params = append(params, fmt.Sprintf("%s=%s", param, value))
		}
		uri += strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	if Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Token))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func Login(email, password string) string {
	loginResp, err := PostApi(LoginAPI.URI, payloadToJson(LoginPayload{
		Email:    email,
		Password: password,
	}))
	if err != nil {
		utils.Fatal("error occurred during login", err.Error())
	}

	var responseJson map[string]interface{}
	err = json.Unmarshal(loginResp, &responseJson)
	if err != nil {
		utils.Fatal("failed to unmarshal login API response", err.Error())
	}

	token, exists := responseJson["token"]
	if !exists {
		utils.Fatal("login failed", "failed to get token from response")
	}

	Token = token.(string)
	if Token == "" {
		utils.Fatal("login failed", "token is unexpectedly empty")
	}
	return Token
}

func PromptLogin() string {
	fmt.Println("enter login credentials.")
	username := utils.GetInput("email")
	password := utils.GetInput("password")

	return Login(username, password)
}
