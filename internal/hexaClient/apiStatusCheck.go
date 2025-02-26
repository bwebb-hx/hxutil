package hexaclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	passColor = color.New(color.FgGreen)
	failColor = color.New(color.FgRed)

	speedFastColor = color.New(color.FgHiGreen)
	speedOkayColor = color.New(color.FgCyan)
	speedPoorColor = color.New(color.FgMagenta)
	speedSlowColor = color.New(color.FgHiRed)
	speedDeadColor = color.New(color.BgRed, color.FgBlack)
)

const (
	// Dear hackers: these credentials are only for testing, so they don't protect anything important.
	TestAccUser   = "b.webb+test@hexabase.com"
	TestAccPass   = "test123"
	TestP_ID      = "674716ff253630d46156a153"
	TestD_ID      = "674724ac4ba983711e015530"
	testAction_ID = "674724acb7eeb7dd909dd15b"
	testVarName   = "ENV_VARIABLE_1"
	testFuncName  = "function1"
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
func testApi(apiDef ApiEndpoint, formatURI []any, queryParams map[string]string, payload interface{}, evalFunc func(data []byte) error, n int, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

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
			resp, err = GetApi(apiDef.URI, queryParams)
		} else if apiDef.Method == POST {
			resp, err = PostApi(apiDef.URI, payloadToJson(payload))
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
	method := apiDef.Method
	if method == "GET" {
		method += " "
	}

	c := failColor
	if fail == 0 {
		c = passColor
	}
	c.Printf("%s %s %s %s ms\n", method, apiDef.DisplayURI, status, speedometer(averageExecTime.Milliseconds()))
}

func speedometer(ms int64) string {
	if ms <= 100 {
		return speedFastColor.Sprint(ms)
	}
	if ms <= 300 {
		return speedOkayColor.Sprint(ms)
	}
	if ms <= 1000 {
		return speedPoorColor.Sprint(ms)
	}
	if ms <= 5000 {
		return speedSlowColor.Sprint(ms)
	}
	return speedDeadColor.Sprint(ms)
}

func login() string {
	return Login(TestAccUser, TestAccPass)
}

// RunStatusCheck tests the connectivity, response time, etc of all APIs (well, those that are registered here so far).
func RunStatusCheck() {
	var wg sync.WaitGroup
	n := 3

	// functions that represent an API test case.
	apis := []func(){
		func() {
			// Login
			testApi(LoginAPI, nil, nil, LoginPayload{
				Email:    TestAccUser,
				Password: TestAccPass,
			}, func(data []byte) error {
				var respJson map[string]interface{}
				if err := json.Unmarshal(data, &respJson); err != nil {
					return err
				}
				if _, exists := respJson["token"]; !exists {
					return errors.New("missing token in response")
				}
				return nil
			}, n, &wg)
		},
		func() {
			// GetWorkspaces
			testApi(GetWorkspacesAPI, nil, nil, nil, func(data []byte) error {
				var respJson map[string]interface{}
				if err := json.Unmarshal(data, &respJson); err != nil {
					return err
				}
				if _, exists := respJson["workspaces"]; !exists {
					return errors.New("missing workspaces in response")
				}
				return nil
			}, n, &wg)
		},
		func() {
			// GetActions
			testApi(GetActionsAPI, []any{TestD_ID}, nil, nil, func(data []byte) error {
				var respJson []map[string]interface{}
				if err := json.Unmarshal(data, &respJson); err != nil {
					return err
				}
				if len(respJson) == 0 {
					return errors.New("no action details found in response")
				}
				return nil
			}, n, &wg)
		},
		func() {
			// DownloadActionScript
			testApi(DownloadActionScriptAPI, []any{testAction_ID}, map[string]string{"script_type": "post"}, nil, func(data []byte) error {
				actionScript := string(data)
				if !strings.Contains(actionScript, "main(data)") {
					return errors.New("failed to load actionscript")
				}
				return nil
			}, n, &wg)
		},
		func() {
			// get env variables
			testApi(GetApplicationScriptVariableAPI, []any{TestP_ID, testVarName}, nil, nil, func(data []byte) error {
				var jsonData map[string]interface{}
				if err := json.Unmarshal(data, &jsonData); err != nil {
					return err
				}
				if val, exists := jsonData["var_name"]; !exists || val != testVarName {
					return errors.New("environment variable name incorrect or missing")
				}
				if val, exists := jsonData["value"]; !exists || val != "secret value" {
					return errors.New("environment variable value incorrect or missing")
				}
				return nil
			}, n, &wg)
		},
		func() {
			// GetDatastores
			testApi(GetDatastoresAPI, []any{TestP_ID}, nil, nil, func(data []byte) error {
				var jsonData []map[string]interface{}
				if err := json.Unmarshal(data, &jsonData); err != nil {
					return err
				}
				if len(jsonData) == 0 {
					return errors.New("no datastores found")
				}
				if _, exists := jsonData[0]["datastore_id"]; !exists {
					return errors.New("datastore_id not found in response")
				}
				return nil
			}, n, &wg)
		},
		func() {
			// (UN) GetFunctionActionScript
			testApi(UN_GetFunctionActionScriptAPI, nil, map[string]string{"p_id": TestP_ID}, nil, func(data []byte) error {
				var jsonData UN_GetFunctionActionScriptResponse
				if err := json.Unmarshal(data, &jsonData); err != nil {
					return err
				}
				if len(jsonData) == 0 {
					return errors.New("response contains no functions")
				}
				function := jsonData[0]
				if function.DisplayID != testFuncName {
					return errors.New("incorrect function name")
				}
				if !strings.Contains(function.Pre.Script, "async function main(data)") {
					return errors.New("function actionscript contents are missing")
				}
				return nil
			}, n, &wg)
		},
		func() {
			// (UN) GetProjectSettings
			testApi(UN_GetProjectSettingsAPI, nil, map[string]string{"p_id": TestP_ID}, nil, func(data []byte) error {
				var jsonData UN_GetProjectSettingsResponse
				if err := json.Unmarshal(data, &jsonData); err != nil {
					return err
				}
				if jsonData.PID != TestP_ID {
					return errors.New("pid is incorrect")
				}
				varFound := false
				for _, scriptVar := range jsonData.ScriptVars {
					if scriptVar.VarName == testVarName {
						varFound = true
						break
					}
				}
				if !varFound {
					return errors.New("env var not found")
				}
				return nil
			}, n, &wg)
		},
	}

	// NoAuth APIs
	testApi(ForgotPasswordAPI, nil, nil, ForgotPasswordPayload{Email: TestAccUser, Host: baseURL}, func(data []byte) error {
		var resp map[string]interface{}
		err := json.Unmarshal(data, &resp)
		if err != nil {
			return err
		}
		val, exists := resp["valid_email"]
		if !exists {
			fmt.Println(resp)
			return errors.New("valid_email key is missing")
		}
		if val != true {
			return errors.New("email is expected to be valid")
		}
		return nil
	}, n, nil)

	// Login to set the auth token for auth APIs
	token := login()
	if token == "" {
		log.Fatal("failed to get token")
	}
	fmt.Println("(login succeeded)")

	// run each api test concurrently
	wg.Add(len(apis))
	for _, fn := range apis {
		go fn()
	}

	wg.Wait()
	fmt.Println("done!")
}
