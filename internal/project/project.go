package project

import (
	"encoding/json"
	"fmt"

	hx "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/bwebb-hx/hxutil/internal/utils"
)

func Diff(p1, p2 string) {
	hx.PromptLogin()

	// diff project settings and env variables
	diffProjectSettings(p1, p2)
	utils.EnterToContinue()

	// diff actionscripts
	diffFunctionActionScripts(p1, p2)
	utils.EnterToContinue()
}

func diffProjectSettings(p1, p2 string) {
	utils.Hint("BEGIN: Diff Project Settings")

	p1Bytes, err := hx.GetApi(hx.UN_GetProjectSettingsAPI.URI, map[string]string{"p_id": p1})
	if err != nil {
		utils.Fatal("failed to get project", err.Error())
	}
	var p1SettingsResponse hx.UN_GetProjectSettingsResponse
	if err = json.Unmarshal(p1Bytes, &p1SettingsResponse); err != nil {
		utils.Fatal("failed to unmarshal json", err.Error())
	}

	p2Bytes, err := hx.GetApi(hx.UN_GetProjectSettingsAPI.URI, map[string]string{"p_id": p2})
	if err != nil {
		utils.Fatal("failed to get project", err.Error())
	}
	var p2SettingsResponse hx.UN_GetProjectSettingsResponse
	if err = json.Unmarshal(p2Bytes, &p2SettingsResponse); err != nil {
		utils.Fatal("failed to unmarshal json", err.Error())
	}

	// compare high level details (names, etc)
	diffValues(p1SettingsResponse.Name.En, p2SettingsResponse.Name.En, "Name (En)")
	diffValues(p1SettingsResponse.Name.Ja, p2SettingsResponse.Name.Ja, "Name (Ja)")
	diffValues(p1SettingsResponse.DisplayID, p2SettingsResponse.DisplayID, "Display ID")

	// compare env vars
	for _, envVar := range p1SettingsResponse.ScriptVars {
		found := false
		for _, envVar2 := range p2SettingsResponse.ScriptVars {
			if envVar.VarName == envVar2.VarName {
				diffValues(envVar.Value, envVar2.Value, envVar.VarName+" (Env)")
				found = true
				break
			}
		}
		if !found {
			utils.Warn("environment variable match not found: "+envVar.VarName, "(exists in p1 but not p2)")
		}
	}
	// confirm that there aren't extra env vars in p2
	for _, envVar := range p2SettingsResponse.ScriptVars {
		found := false
		for _, envVar2 := range p1SettingsResponse.ScriptVars {
			if envVar.VarName == envVar2.VarName {
				found = true
				break
			}
		}
		if !found {
			utils.Warn("environment variable match not found: "+envVar.VarName, "(exists in p2 but not p1)")
		}
	}

	utils.Hint("END: Diff Project Settings")
}

func diffValues(p1Val, p2Val string, valueName string) {
	if p1Val != p2Val {
		fmt.Println("\nDiff found!:", valueName)

		if len(p1Val) > 50 && len(p2Val) > 50 {
			utils.Hint("(Showing diff since values are large)")
			fmt.Println(utils.GetDiff(p1Val, p2Val))
			return
		}

		fmt.Println("p1:", p1Val)
		fmt.Println("p2:", p2Val)
	}
}

func diffFunctionActionScripts(p1, p2 string) {
	utils.Hint("BEGIN: Diff Project Functions")
	// get action scripts for functions
	p1FnBytes, err := hx.GetApi(hx.UN_GetFunctionActionScriptAPI.URI, map[string]string{
		"p_id": p1,
	})
	if err != nil {
		utils.Fatal("failed to get functions", err.Error())
	}
	var p1Functions hx.UN_GetFunctionActionScriptResponse
	if err = json.Unmarshal(p1FnBytes, &p1Functions); err != nil {
		utils.Fatal("failed to unmarshal", err.Error())
	}

	p2FnBytes, err := hx.GetApi(hx.UN_GetFunctionActionScriptAPI.URI, map[string]string{
		"p_id": p2,
	})
	if err != nil {
		utils.Fatal("failed to get functions", err.Error())
	}
	var p2Functions hx.UN_GetFunctionActionScriptResponse
	if err = json.Unmarshal(p2FnBytes, &p2Functions); err != nil {
		utils.Fatal("failed to unmarshal", err.Error())
	}

	// diff function actionscripts
	diffFunctions := []string{}
	for _, function := range p1Functions {
		found := false
		for _, function2 := range p2Functions {
			if function.DisplayID == function2.DisplayID {
				// match found; diff contents
				found = true
				diff := utils.GetDiff(function.Pre.Script, function2.Pre.Script)
				if diff != "" {
					diffFunctions = append(diffFunctions, function.DisplayID)
					fmt.Println("\n==========")
					fmt.Println("Diff found!:", function.DisplayID, "(Function)")
					fmt.Println(diff)
					fmt.Println("==========")
					utils.EnterToContinue()
				}
				break
			}
		}
		if !found {
			diffFunctions = append(diffFunctions, function.DisplayID+" (MISSING in p2)")
			utils.Warn("Function match not found: "+function.DisplayID, "(found in p1, but not p2)")
		}
	}
	// make sure there aren't functions in p2 that aren't in p1
	for _, function := range p2Functions {
		found := false
		for _, function2 := range p1Functions {
			if function.DisplayID == function2.DisplayID {
				found = true
				break
			}
		}
		if !found {
			diffFunctions = append(diffFunctions, function.DisplayID+" (MISSING in p1)")
			utils.Warn("Function match not found: "+function.DisplayID, "(found in p2, but not p1)")
		}
	}

	fmt.Println("=== Summary ===")
	for _, fn := range diffFunctions {
		fmt.Println(fn)
	}

	utils.Hint("END: Diff Project Functions")
}
