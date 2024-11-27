package hexaclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var baseURL = "https://api.hexabase.com"

func SetBaseUrl(url string) {
	baseURL = url
}

func PostApi(uri string, body []byte) ([]byte, error) {
	resp, err := http.Post(fmt.Sprintf("%s%s", baseURL, uri), "application/json", bytes.NewReader(body))
	if err != nil {
		return []byte{}, err
	}
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
	resp, err := http.Get(fmt.Sprintf("%s%s", baseURL, uri))
	if err != nil {
		return []byte{}, err
	}
	return io.ReadAll(resp.Body)
}
