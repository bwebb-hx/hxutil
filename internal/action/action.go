package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwebb-hx/hxutil/internal/config"
	hexaclient "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/bwebb-hx/hxutil/internal/utils"
)

const OUT = ".as_temp"

var INTERACTIVE_MODE = true

type Action struct {
	ID            string
	DisplayID     string
	Name          string
	P_ID          string
	D_ID          string
	DatastoreName string
}

func getAllActionIDs(d_id string, datastoreName string) []Action {
	resp, err := hexaclient.GetApi(fmt.Sprintf(hexaclient.GetActionsAPI.URI, d_id), nil)
	if err != nil {
		log.Println("failed to get action IDs:", err)
		return nil
	}
	var getActionsResp hexaclient.GetActionsResponse
	if err := json.Unmarshal(resp, &getActionsResp); err != nil {
		log.Println("failed to unmarshal getActions API response")
		return nil
	}

	actions := make([]Action, 0)
	for _, actionDef := range getActionsResp {
		action := Action{
			ID:            actionDef.ActionID,
			DisplayID:     actionDef.DisplayID,
			Name:          actionDef.Name,
			P_ID:          actionDef.P_ID,
			D_ID:          actionDef.D_ID,
			DatastoreName: datastoreName,
		}
		actions = append(actions, action)
	}

	return actions
}

func GetProjectActions(p_id string) []Action {
	// get all actionscripts IDs for all datastores in the project
	getDatastoresBytes, err := hexaclient.GetApi(fmt.Sprintf(hexaclient.GetDatastoresAPI.URI, p_id), nil)
	if err != nil {
		log.Fatal("failed to get datastores for project:", err)
	}
	var datastores hexaclient.GetDatastoresResponse
	if err := json.Unmarshal(getDatastoresBytes, &datastores); err != nil {
		log.Fatal("failed to unmarshal datastores response:", err)
	}
	if len(datastores) == 0 {
		utils.Warn("No datastores found in project", "P_ID: "+p_id)
		fmt.Println(string(getDatastoresBytes))
		return []Action{}
	}

	actions := make([]Action, 0)
	for _, datastore := range datastores {
		actions = append(actions, getAllActionIDs(datastore.DatastoreID, datastore.Name)...)
	}

	return actions
}

func DiffActionScripts(absPath string) {
	// load config
	c := config.GetConfig()
	project := c.SelectProject()
	if project == nil {
		utils.Fatal("no project selected", "")
	}

	// login to hexabase
	c.SelectUserAndLogin(project.P_ID)

	// get all actionscripts IDs for all datastores in the project
	actions := GetProjectActions(project.P_ID)
	if len(actions) == 0 {
		log.Fatal("No actions found in the given project:", project.P_ID)
	}

	// get all function actionscripts in the project
	getFunctionsBytes, err := hexaclient.GetApi(hexaclient.UN_GetFunctionActionScriptAPI.URI, map[string]string{
		"p_id": project.P_ID,
	})
	if err != nil {
		log.Fatal("failed to get functions for project:", err)
	}

	var functions hexaclient.UN_GetFunctionActionScriptResponse
	if err := json.Unmarshal(getFunctionsBytes, &functions); err != nil {
		log.Println("failed to get actionscripts for functions")
	}

	diffFiles := make([]string, 0)
	totalComps := 0
	diffSearchErrs := diffSearchErrs{}

	for _, action := range actions {
		// find pre scripts
		diff, searchErrs := diffActionScript(action, absPath, "pre")
		if searchErrs.errOccurred() {
			diffSearchErrs.combineCounts(searchErrs)

			if searchErrs.localNotFound > 0 {
				diffFiles = append(diffFiles, fmt.Sprintf("**LOCAL NOT FOUND**: %s (pre) [%s]", action.DisplayID, action.DatastoreName))
			}
		}
		if diff {
			diffFiles = append(diffFiles, fmt.Sprintf("%s (pre) [%s]", action.DisplayID, action.DatastoreName))
		}

		// find post scripts
		diff, searchErrs = diffActionScript(action, absPath, "post")
		if searchErrs.errOccurred() {
			diffSearchErrs.combineCounts(searchErrs)

			if searchErrs.localNotFound > 0 {
				diffFiles = append(diffFiles, fmt.Sprintf("**LOCAL NOT FOUND**: %s (post) [%s]", action.DisplayID, action.DatastoreName))
			}
		}
		if diff {
			diffFiles = append(diffFiles, fmt.Sprintf("%s (post) [%s]", action.DisplayID, action.DatastoreName))
		}
		totalComps++
	}

	diffFunctions, diffFnSearchErrs := diffFunctionActionScripts(absPath, functions)
	diffFiles = append(diffFiles, diffFunctions...)
	if diffFnSearchErrs.errOccurred() {
		diffSearchErrs.combineCounts(diffFnSearchErrs)
	}

	fmt.Println("\nSUMMARY\n=======")
	fmt.Print("ActionScripts with local differences:\n\n")
	for _, diffFile := range diffFiles {
		fmt.Println(diffFile)
	}
	fmt.Println("\nTotal files checked:", totalComps)
	if diffSearchErrs.errOccurred() {
		fmt.Println(diffSearchErrs)
	}
	fmt.Println("=======\n ")
}

type diffSearchErrs struct {
	localNotFound  int
	respUnexpected int
	walkDirErr     int
}

func (dse diffSearchErrs) String() string {
	out := fmt.Sprintf("%s: %v\n", "local scripts not found", dse.localNotFound)
	out += fmt.Sprintf("%s: %v\n", "unexpected API responses", dse.respUnexpected)
	out += fmt.Sprintf("%s: %v", "errors while walking project files", dse.walkDirErr)
	return out
}

func (dse diffSearchErrs) errOccurred() bool {
	return dse.localNotFound > 0 || dse.respUnexpected > 0 || dse.walkDirErr > 0
}

func (dse *diffSearchErrs) combineCounts(searchErrs diffSearchErrs) {
	dse.localNotFound += searchErrs.localNotFound
	dse.respUnexpected += searchErrs.respUnexpected
	dse.walkDirErr += searchErrs.walkDirErr
}

func diffFunctionActionScripts(absPath string, functions hexaclient.UN_GetFunctionActionScriptResponse) ([]string, diffSearchErrs) {
	diffFiles := make([]string, 0)
	searchErrs := diffSearchErrs{}

	for _, function := range functions {
		actionscript := strings.TrimSpace(function.Pre.Script)
		if actionscript == "" {
			log.Println("empty function?:", function.DisplayID)
			continue
		}

		// find the corresponding file
		stop := errors.New("stop")
		found := false
		diffVal := false
		err := filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasPrefix(d.Name(), function.DisplayID) {
				if strings.HasSuffix(d.Name(), ".js") {
					// match found! get diff results
					diffVal = diff(path, actionscript, d.Name(), "FUNCTION")
					found = true
					return stop
				}
			}
			return nil
		})
		if err != nil && !errors.Is(err, stop) {
			log.Println("an error occurred while walking project files:", err)
			searchErrs.walkDirErr++
			continue
		}
		if !found {
			log.Println("failed to find function:", function.DisplayID)
			searchErrs.localNotFound++
			diffFiles = append(diffFiles, fmt.Sprintf("**LOCAL NOT FOUND**: %s [FUNCTION]", function.DisplayID))
			continue
		}
		if diffVal {
			diffFiles = append(diffFiles, function.DisplayID+" [FUNCTION]")
		}
	}

	return diffFiles, searchErrs
}

func DownloadActionScript(actionID string, scriptType string) (string, error) {
	downloadResp, err := hexaclient.GetApi(fmt.Sprintf(hexaclient.DownloadActionScriptAPI.URI, actionID), map[string]string{
		"script_type": scriptType,
	})
	if err != nil {
		return "", err
	}
	actionscript := strings.TrimSpace(string(downloadResp))
	if actionscript == "" {
		return "", nil
	}

	// check for errors
	if actionscript[0] == '{' {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(downloadResp, &jsonData); err != nil {
			utils.Error("tried to unmarshal actionscript json response but error occurred", err.Error())
			return "", errors.New("failed to unmarshal as json; unexpected response format")
		}
		errorCode, exists := jsonData["error_code"]
		if exists {
			if errorCode == "NOT_FOUND" {
				return "", nil
			}
			if errorCode == "SYSTEM_ERROR" {
				if jsonData["error"] == "empty script" {
					return "", nil
				}
			}
		}
		utils.Warn("unexpected actionscript download response:", actionscript)
		return "", errors.New("unexpected actionscript download response")
	}

	return actionscript, nil
}

func diffActionScript(action Action, absPath string, scriptType string) (bool, diffSearchErrs) {
	stats := diffSearchErrs{}
	diffVal := false

	if scriptType != "post" && scriptType != "pre" {
		utils.Error("unsupported script type: "+scriptType, "")
		return false, stats
	}

	// download actionscript
	actionscript, err := DownloadActionScript(action.ID, scriptType)
	if err != nil {
		utils.Error("error occurred while fetching actionscript", err.Error())
		stats.respUnexpected++
		return false, stats
	}
	if actionscript == "" {
		return false, stats
	}

	stop := errors.New("stop")
	found := false
	err = filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), action.DisplayID) {
			if strings.HasSuffix(d.Name(), scriptType+".js") {
				// match found! get diff results
				diffVal = diff(path, actionscript, d.Name(), action.DatastoreName)
				found = true
				return stop
			}
		}
		return nil
	})
	if err != nil && !errors.Is(err, stop) {
		log.Println("an error occurred while walking project files:", err)
		stats.walkDirErr++
	}
	if !found {
		log.Println("failed to find actionscript in local:", action.Name, fmt.Sprintf("(%s)", action.DatastoreName))
		if len(actionscript) > 15 {
			fmt.Println("actionscript snippet:", actionscript[:10]+"...")
		} else {
			log.Println("actionscript snippet:", actionscript)
		}
		stats.localNotFound++
	}

	return diffVal, stats
}

// returns true if a difference is found
func diff(local, remoteString, fileName, datastoreName string) bool {
	localBytes, err := os.ReadFile(local)
	if err != nil {
		log.Println("failed to load local file:", err)
		return false
	}

	diff := utils.GetDiff(string(localBytes), remoteString)
	if diff != "" {
		fmt.Println("\n===")
		fmt.Println(fileName, fmt.Sprintf("(%s)\n", datastoreName))
		fmt.Println(diff)
		fmt.Println("===")
		if INTERACTIVE_MODE {
			utils.EnterToContinue()
		}
		return true
	}
	return false
}
