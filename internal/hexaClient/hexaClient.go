package hexaclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var baseURL = "https://api.hexabase.com"

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

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
