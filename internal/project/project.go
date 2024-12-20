package project

import (
	"encoding/json"
	"fmt"
	"log"

	hx "github.com/bwebb-hx/hxutil/internal/hexaClient"
)

func Diff(p1, p2 string) {
	hx.PromptLogin()
	// diff project settings and env variables
	diffProjectSettings(p1, p2)
}

func diffProjectSettings(p1, p2 string) {
	p1Bytes, err := hx.GetApi(hx.UN_GetProjectSettingsAPI.URI, map[string]string{"p_id": p1})
	if err != nil {
		log.Fatal("failed to get project:", err)
	}
	var p1SettingsResponse hx.UN_GetProjectSettingsResponse
	if err = json.Unmarshal(p1Bytes, &p1SettingsResponse); err != nil {
		log.Fatal("failed to unmarshal json:", err)
	}

	p2Bytes, err := hx.GetApi(hx.UN_GetProjectSettingsAPI.URI, map[string]string{"p_id": p2})
	if err != nil {
		log.Fatal("failed to get project:", err)
	}
	var p2SettingsResponse hx.UN_GetProjectSettingsResponse
	if err = json.Unmarshal(p2Bytes, &p2SettingsResponse); err != nil {
		log.Fatal("failed to unmarshal json:", err)
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
			fmt.Println("\n**WARNING: environment variable match not found:", envVar.VarName)
			fmt.Println("  (exists in p1 but not p2)")
		}
	}
}

func diffValues(p1Val, p2Val any, valueName string) {
	if p1Val != p2Val {
		fmt.Println("\nDiff found!:", valueName)
		fmt.Println("p1:", p1Val)
		fmt.Println("p2:", p2Val)
	}
}
