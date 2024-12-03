package hexaclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	// Dear hackers: these credentials are only for testing, so they don't protect anything important.
	testAccUser = "b.webb+test@hexabase.com"
	testAccPass = "test123"
	testP_ID    = "674716ff253630d46156a153"
	testD_ID    = "674724ac4ba983711e015530"
)

const (
	// signal to just check that a key exists, and not worry about the specific value
	EXISTS_CHECK = "<<EXISTS>>"

	// methods
	GET  = "GET"
	POST = "POST"
)

func payloadToJson(data interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal("error converting payload to json:", err)
	}
	return jsonData
}

// testApi tests the given API with the given payload.
func testApi(apiDef ApiEndpoint, formatURI []any, queryParams map[string]string, payload interface{}, evalFunc func(data []byte) error, token string, n int, wg *sync.WaitGroup) {
	defer wg.Done()

	if formatURI != nil {
		apiDef.URI = fmt.Sprintf(apiDef.URI, formatURI...)
	}

	var totalTime time.Duration
	pass, fail := 0, 0

	if n == 0 {
		fmt.Println(apiDef.URI, "(n = 0; abort)")
		return
	}

	for i := 0; i < n; i++ {
		start := time.Now()
		var resp []byte
		var err error
		if apiDef.Method == GET {
			resp, err = GetApi(apiDef.URI, queryParams, token)
		} else if apiDef.Method == POST {
			resp, err = PostApi(apiDef.URI, payloadToJson(payload), token)
		} else {
			log.Println("Error: unknown HTTP method", apiDef.Method)
			return
		}
		if err != nil {
			log.Println("failed to call API:", err)
			return
		}
		totalTime += time.Since(start)

		err = evalFunc(resp)

		if err != nil {
			log.Println(apiDef.URI, err)
			fail++
		} else {
			pass++
		}
	}

	averageExecTime := totalTime / time.Duration(n)
	status := fmt.Sprintf("%v/%v", pass, n)
	if fail > 0 {
		status += " (FAIL)"
	} else {
		status += " (Pass!)"
	}
	fmt.Printf("%s %s %d ms\n", apiDef.URI, status, averageExecTime.Milliseconds())
}

func login() string {
	loginResp, err := PostApi(LoginAPI.URI, payloadToJson(LoginPayload{
		Email:    testAccUser,
		Password: testAccPass,
	}), "")
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
	return token.(string)
}

// RunStatusCheck tests the connectivity, response time, etc of all APIs (well, those that are registered here so far).
func RunStatusCheck() {
	var wg sync.WaitGroup
	wg.Add(3)
	n := 3

	// First, officially login to get the token
	token := login()
	if token == "" {
		log.Fatal("failed to get token")
	}
	fmt.Println("(login succeeded)")

	// Login
	go testApi(LoginAPI, nil, nil, LoginPayload{
		Email:    testAccUser,
		Password: testAccPass,
	}, func(data []byte) error {
		var respJson map[string]interface{}
		if err := json.Unmarshal(data, &respJson); err != nil {
			return err
		}
		if _, exists := respJson["token"]; !exists {
			return errors.New("missing token in response")
		}
		return nil
	}, "", n, &wg)

	// Workspaces
	go testApi(GetWorkspacesAPI, nil, nil, nil, func(data []byte) error {
		var respJson map[string]interface{}
		if err := json.Unmarshal(data, &respJson); err != nil {
			return err
		}
		if _, exists := respJson["workspaces"]; !exists {
			return errors.New("missing workspaces in response")
		}
		return nil
	}, token, n, &wg)

	// GetActions
	go testApi(GetActionsAPI, []any{testD_ID}, nil, nil, func(data []byte) error {
		var respJson []map[string]interface{}
		if err := json.Unmarshal(data, &respJson); err != nil {
			return err
		}
		if len(respJson) == 0 {
			return errors.New("no action details found in response")
		}
		return nil
	}, token, n, &wg)

	wg.Wait()
	fmt.Println("done!")
}
