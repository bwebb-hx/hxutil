package hexaclient

import (
	"encoding/json"
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
func testApi(uri string, method string, queryParams map[string]string, payload interface{}, expOut map[string]interface{}, n int, wg *sync.WaitGroup) {
	defer wg.Done()

	var totalTime time.Duration
	pass, fail := 0, 0

	for i := 0; i < n; i++ {
		start := time.Now()
		var resp []byte
		var err error
		if method == GET {
			resp, err = GetApi(uri, queryParams)
		} else if method == POST {
			resp, err = PostApi(uri, payloadToJson(payload))
		} else {
			log.Println("Error: unknown HTTP method", method)
			return
		}
		if err != nil {
			log.Println("failed to call API:", err)
			return
		}
		totalTime += time.Since(start)

		var responseJson map[string]interface{}
		err = json.Unmarshal(resp, &responseJson)
		if err != nil {
			log.Println("failed to unmarshal API response:", err)
			return
		}

		// check that all expected values in expOut exist and match in responseJson
		badResponse := false
		for key, val := range expOut {
			respVal, exists := responseJson[key]
			if !exists {
				log.Println("Error: expected data not found in API response.")
				log.Println("Expected key:", key)
				badResponse = true
				continue
			}
			if val == EXISTS_CHECK {
				// value exists, so we good
				continue
			}
			if respVal != val {
				log.Println("Error: API response has data that doesn't match the expected output.")
				log.Printf("(key=%s) Expected: %s, Got: %s\n", key, val, respVal)
				badResponse = true
			}
		}

		if badResponse {
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
	fmt.Printf("%s %s %d ms\n", uri, status, averageExecTime.Milliseconds())
}

// RunStatusCheck tests the connectivity, response time, etc of all APIs (well, those that are registered here so far).
func RunStatusCheck() {
	var wg sync.WaitGroup
	wg.Add(2)

	n := 3
	// Login
	go testApi(LoginURI, POST, nil, LoginPayload{
		Email:    testAccUser,
		Password: testAccPass,
	}, map[string]interface{}{
		"token": EXISTS_CHECK,
	}, n, &wg)
	// GetActions
	go testApi(fmt.Sprintf(GetActionsURI, testD_ID), GET, nil, nil, nil, n, &wg)

	wg.Wait()
	fmt.Println("done!")
}
