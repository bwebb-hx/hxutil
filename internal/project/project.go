package project

import (
	"encoding/json"
	"fmt"

	"github.com/bwebb-hx/hxutil/internal/action"
	"github.com/bwebb-hx/hxutil/internal/config"
	hx "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/bwebb-hx/hxutil/internal/utils"
)

func Diff(p1, p2 string) {
	c := config.GetConfig()
	if c == nil {
		hx.PromptLogin()
	} else {
		c.SelectUserAndLogin(p1)
	}

	// if no projects provided, use config and prompt user
	if p1 == "" {
		utils.Hint("Select PID 1")
		project1 := c.SelectProject()
		if project1 == nil {
			utils.Fatal("failed to select project", "project ID required for this utility")
		}
		p1 = project1.P_ID
	}
	if p2 == "" {
		utils.Hint("Select PID 2")
		project2 := c.SelectProject()
		if project2 == nil {
			utils.Fatal("failed to select project", "project ID required for this utility")
		}
		p2 = project2.P_ID
	}

	// diff project settings and env variables
	diffProjectSettings(p1, p2)
	utils.EnterToContinue()

	// diff functions
	diffFunctionActionScripts(p1, p2)
	utils.EnterToContinue()

	// diff actionscripts
	diffDatastoreActionScripts(p1, p2)
	utils.EnterToContinue()
}

func diffProjectSettings(p1, p2 string) {
	utils.Hint("Diffing Project Settings...")

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

	utils.Hint(fmt.Sprintf("p1: %s [%s]", p1SettingsResponse.DisplayID, p1SettingsResponse.PID))
	utils.Hint(fmt.Sprintf("p2: %s [%s]", p2SettingsResponse.DisplayID, p2SettingsResponse.PID))

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
	utils.Hint("Diffing Project Functions...")
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
	diffLogs := []string{}
	for _, function := range p1Functions {
		found := false
		for _, function2 := range p2Functions {
			if function.DisplayID == function2.DisplayID {
				// match found; diff contents
				found = true

				if function.Pre.Script == "" || function2.Pre.Script == "" {
					if function.Pre.Script != "" {
						utils.ColorError.Println("!!MISSING: Function script defined in p1 but not p2")
						diffLogs = append(diffLogs, fmt.Sprintf("%s %s (Function)", utils.ColorError.Sprint("MISSING IN P2:"), function.DisplayID))
						break
					}
					if function2.Pre.Script != "" {
						utils.ColorError.Println("!!MISSING: Function script defined in p2 but not p1")
						diffLogs = append(diffLogs, fmt.Sprintf("%s %s (Function)", utils.ColorError.Sprint("MISSING IN P1:"), function.DisplayID))
						break
					}
					// both are empty?
					utils.ColorError.Printf("Empty? Function %s in both projects is empty...\n", function.DisplayID)
					diffLogs = append(diffLogs, utils.ColorError.Sprintf("Empty Functions: %s in both projects is empty...", function.DisplayID))
					break
				}

				diff := utils.GetDiff(function.Pre.Script, function2.Pre.Script)
				if diff != "" {
					diffLogs = append(diffLogs, fmt.Sprintf("%s %s (Function)", utils.ColorWarn.Sprint("DIFF FOUND:"), function.DisplayID))
					utils.ColorWarn.Println("\nDiff Found!", function.DisplayID, "(Function)")
					fmt.Println(diff)
					utils.Hint("(End Diff)")
					utils.EnterToContinue()
				}
				break
			}
		}
		if !found {
			diffLogs = append(diffLogs, fmt.Sprintf("? NO MATCH: Function %s found in P1 but not P2.", function.DisplayID))
			diffLogs = append(diffLogs, utils.ColorHint.Sprint("  Confirm function exists and display IDs match between projects"))
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
			diffLogs = append(diffLogs, fmt.Sprintf("? NO MATCH: Function %s found in P2 but not P1.", function.DisplayID))
			diffLogs = append(diffLogs, utils.ColorHint.Sprint("  Confirm function exists and display IDs match between projects"))
		}
	}

	fmt.Println("\n== SUMMARY ==")
	for _, fn := range diffLogs {
		fmt.Println(fn)
	}
}

func diffDatastoreActionScripts(p1, p2 string) {
	utils.Hint("Diffing Datastore ActionScripts...")

	p1Actions := action.GetProjectActions(p1)
	p2Actions := action.GetProjectActions(p2)
	if len(p1Actions) == 0 {
		utils.Hint("(No actions found for p1)")
	}
	if len(p2Actions) == 0 {
		utils.Hint("(No actions found for p2)")
	}
	if len(p1Actions) == 0 || len(p2Actions) == 0 {
		return
	}

	// find matching actions and diff them
	// an action matches if it has the same display ID, and the same datastore name
	diffLog := make([]string, 0)
	for _, action1 := range p1Actions {
		found := false
		for _, action2 := range p2Actions {
			if action1.DisplayID == action2.DisplayID {
				if action1.DatastoreName != action2.DatastoreName {
					continue
				}
				found = true

				diffScripts := func(scriptType string) {
					// download actionscripts
					script1, err := action.DownloadActionScript(action1.ID, scriptType)
					if err != nil {
						utils.Error("error while downloading actionscript", err.Error())
						return
					}
					script2, err := action.DownloadActionScript(action2.ID, scriptType)
					if err != nil {
						utils.Error("error while downloading actionscript", err.Error())
						return
					}
					if script1 == "" || script2 == "" {
						// one is empty, but not the other
						if script1 != "" || script2 != "" {
							if script1 != "" {
								utils.ColorError.Println("!!MISSING: ActionScript defined in p1 but not p2")
								diffLog = append(diffLog, fmt.Sprintf("%s %s (%s) [%s]", utils.ColorError.Sprint("MISSING IN P2:"), action1.DisplayID, scriptType, action1.DatastoreName))
							} else {
								utils.ColorError.Println("!!MISSING: ActionScript defined in p2 but not p1")
								diffLog = append(diffLog, fmt.Sprintf("%s %s (%s) [%s]", utils.ColorError.Sprint("MISSING IN P1:"), action1.DisplayID, scriptType, action1.DatastoreName))
							}
							utils.ColorWarn.Printf("Action: %s (%s)  Datastore: %s\n", action1.Name, scriptType, action1.DatastoreName)
						}
						return
					}

					// diff
					diff := utils.GetDiff(script1, script2)
					if diff != "" {
						utils.ColorWarn.Println("\nDiff Found!")
						utils.ColorWarn.Printf("Action: %s (%s)  Datastore: %s\n", action1.DisplayID, scriptType, action1.DatastoreName)
						fmt.Println(diff)
						utils.Hint("(End Diff)")

						utils.EnterToContinue()

						diffLog = append(diffLog, fmt.Sprintf("%s %s (%s) [%s]", utils.ColorWarn.Sprint("DIFF FOUND:"), action1.DisplayID, scriptType, action1.DatastoreName))
					}
				}

				diffScripts("pre")
				diffScripts("post")
			}
		}

		if !found {
			diffLog = append(diffLog, fmt.Sprintf("? NO MATCH: Action %s [%s] found in P1 but not P2.", action1.DisplayID, action1.DatastoreName))
			diffLog = append(diffLog, utils.ColorHint.Sprint("  Confirm action exists and display IDs (action and datastore) match between projects"))
		}
	}

	// make sure p2 doesn't have extra actions
	for _, action2 := range p2Actions {
		found := false
		for _, action1 := range p1Actions {
			if action2.DisplayID == action1.DisplayID && action2.DatastoreName == action1.DatastoreName {
				found = true
				break
			}
		}
		if !found {
			diffLog = append(diffLog, fmt.Sprintf("? NO MATCH: Action %s [%s] found in P2 but not P1.", action2.DisplayID, action2.DatastoreName))
			diffLog = append(diffLog, utils.ColorHint.Sprint("  Confirm action exists and display IDs (action and datastore) match between projects"))
		}
	}

	// print summary
	fmt.Println("\n== SUMMARY ==")
	for _, s := range diffLog {
		fmt.Println(s)
	}
}
