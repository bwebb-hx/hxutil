package hexaclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", baseURL, uri), bytes.NewReader(body))
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
	if queryParams != nil {
		params := make([]string, 0)
		uri += "?"
		for param, value := range queryParams {
			params = append(params, fmt.Sprintf("%s=%s", param, value))
		}
		uri += strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", baseURL, uri), nil)
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
		log.Fatal("failed to login with test user:", err)
	}

	var responseJson map[string]interface{}
	err = json.Unmarshal(loginResp, &responseJson)
	if err != nil {
		log.Fatal("failed to unmarshal API response:", err)
	}

	token, exists := responseJson["token"]
	if !exists {
		log.Fatal("failed to get token from response")
	}

	Token = token.(string)
	return Token
}
