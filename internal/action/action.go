package action

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	hexaclient "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/bwebb-hx/hxutil/internal/utils"
)

const OUT = ".as_temp"
const CONFIG = ".hx-tools.json"

type configData struct {
	Username       string `json:"username"`
	Password       string
	P_ID           string `json:"p_id"`
	unsavedChanges bool
}

type action struct {
	ID            string
	DisplayID     string
	Name          string
	P_ID          string
	D_ID          string
	DatastoreName string
}

func getAllActionIDs(d_id string, datastores hexaclient.GetDatastoresResponse) []action {
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

	actions := make([]action, 0)
	for _, actionDef := range getActionsResp {
		action := action{
			ID:        actionDef.ActionID,
			DisplayID: actionDef.DisplayID,
			Name:      actionDef.Name,
			P_ID:      actionDef.P_ID,
			D_ID:      actionDef.D_ID,
		}
		// find the name of the datastore
		for _, ds := range datastores {
			if ds.DatastoreID == action.D_ID {
				action.DatastoreName = ds.Name
				break
			}
		}
		actions = append(actions, action)
	}

	return actions
}

func DiffActionScripts(absPath string) {
	// load config, if it exists
	config := loadConfig()

	// login to hexabase
	token := hexaclient.Login(config.Username, config.Password)
	if token == "" {
		log.Fatal("failed to get login token")
	}

	// get all actionscripts IDs for all datastores in the project
	getDatastoresBytes, err := hexaclient.GetApi(fmt.Sprintf(hexaclient.GetDatastoresAPI.URI, config.P_ID), nil)
	if err != nil {
		log.Fatal("failed to get datastores for project:", err)
	}
	var datastores hexaclient.GetDatastoresResponse
	if err := json.Unmarshal(getDatastoresBytes, &datastores); err != nil {
		log.Fatal("failed to unmarshal datastores response:", err)
	}

	actions := make([]action, 0)
	for _, datastore := range datastores {
		actions = append(actions, getAllActionIDs(datastore.DatastoreID, datastores)...)
	}
	if len(actions) == 0 {
		log.Fatal("No actions found in the given project:", config.P_ID)
	}

	diffFiles := make([]string, 0)
	totalComps := 0
	diffSearchErrs := diffSearchErrs{}

	for _, action := range actions {
		// find pre scripts
		diff, searchErrs := diffActionScript(action, absPath, "pre")
		if searchErrs.errOccurred() {
			diffSearchErrs.combineCounts(searchErrs)
		}
		if diff {
			diffFiles = append(diffFiles, action.DisplayID+" (pre)")
		}

		// find post scripts
		diff, searchErrs = diffActionScript(action, absPath, "post")
		if searchErrs.errOccurred() {
			diffSearchErrs.combineCounts(searchErrs)
		}
		if diff {
			diffFiles = append(diffFiles, action.DisplayID+" (post)")
		}
		totalComps++
	}

	fmt.Println("\nSUMMARY\n=======")
	fmt.Println("ActionScripts with local differences:", diffFiles)
	fmt.Println("Total files checked:", totalComps)
	if diffSearchErrs.errOccurred() {
		fmt.Println(diffSearchErrs)
	}
	fmt.Println("=======\n ")

	if config.unsavedChanges && strings.ToLower(getInput("Save project/user details for next time? [Y/n]")) == "y" {
		saveConfig(config)
	}
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

func diffActionScript(action action, absPath string, scriptType string) (bool, diffSearchErrs) {
	stats := diffSearchErrs{}
	diffVal := false

	if scriptType != "post" && scriptType != "pre" {
		log.Println("unsupported script type:", scriptType)
		return false, stats
	}

	// download actionscript
	downloadResp, err := hexaclient.GetApi(fmt.Sprintf(hexaclient.DownloadActionScriptAPI.URI, action.ID), map[string]string{
		"script_type": scriptType,
	})
	if err != nil {
		log.Println("failed to load actionscript:", err)
		return false, stats
	}
	actionscript := strings.TrimSpace(string(downloadResp))
	if actionscript == "" {
		log.Println("no actionscript found")
		return false, stats
	}
	if actionscript[0] == '{' {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(downloadResp, &jsonData); err != nil {
			log.Println("tried to unmarshal actionscript json response but error occurred:", err)
			stats.respUnexpected++
			return false, stats
		}
		errorCode, exists := jsonData["error_code"]
		if exists {
			if errorCode == "NOT_FOUND" {
				return false, stats
			}
			if errorCode == "SYSTEM_ERROR" {
				if jsonData["error"] == "empty script" {
					return false, stats
				}
			}
		}
		log.Println("unexpected actionscript download response:", actionscript)
		stats.respUnexpected++
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
				diffVal = diff(path, actionscript, d.Name())
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
func diff(local, remoteString, fileName string) bool {
	localBytes, err := os.ReadFile(local)
	if err != nil {
		log.Println("failed to load local file:", err)
		return false
	}
	diff := utils.GetDiff(string(localBytes), remoteString, utils.DiffParams{TrimEqual: true})
	if diff != "" {
		fmt.Println("\n===")
		fmt.Println(fileName)
		fmt.Println(diff)
		fmt.Println("===")
		return true
	}
	return false
}

func getInput(prompt string) string {
	var input string
	fmt.Print(prompt, ": ")
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Fatal("failed to read input:", err)
	}
	return input
}

func saveConfig(config configData) {
	bytes, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Println("failed to save config json:", err)
		return
	}

	err = os.WriteFile(CONFIG, bytes, 0644)
	if err != nil {
		log.Println("failed to write json:", err)
		return
	}

	var gitignore *os.File
	if _, err := os.Stat(CONFIG); os.IsNotExist(err) {
		// gitignore doesn't exist yet (create)
		gitignore, err = os.Create(".gitignore")
		if err != nil {
			log.Println("failed to create .gitignore:", err)
			return
		}
	} else {
		// gitignore already exists
		gitignore, err = os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Println("failed to open .gitignore:", err)
			return
		}
	}
	defer gitignore.Close()

	// check if gitignore is already set
	scanner := bufio.NewScanner(gitignore)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, CONFIG) {
			// already set; no changes needed
			fmt.Printf("Created config file %s.\nNOTE: do not commit this config file to your git repo! It may contain sensitive login information.\n", CONFIG)
			fmt.Println("Your gitignore file should already have an exclusion for this file.")
			return
		}
	}

	// add config file to gitignore
	_, err = gitignore.WriteString("\n\n# ActionScript checker utility\n" + CONFIG + "\n")
	if err != nil {
		log.Println("failed to write to gitignore:", err)
		return
	}

	fmt.Printf("Created config file %s.\nNOTE: do not commit this config file to your git repo! It may contain sensitive login information.\n", CONFIG)
	fmt.Println("Your gitignore file was updated to add an exclusion for this file.")
}

func loadConfig() configData {
	data, err := os.ReadFile(CONFIG)
	if err != nil {
		fmt.Println("(No existing config found)")
		return getUserConfig()
	}

	var config configData
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Println("failed to unmarshal config json:", err)
		return configData{}
	}
	fmt.Println("Existing config found")
	fmt.Println("=====================")
	fmt.Println("Username:", config.Username)
	fmt.Println("Project ID:", config.P_ID)
	if strings.ToLower(getInput("\n\nUse above config? [y/n]")) != "y" {
		fmt.Println("ignoring existing config.")
		return getUserConfig()
	}

	// get password, since it's not saved in the config file
	config.Password = getInput("login password")

	return config
}

func getUserConfig() configData {
	fmt.Println("Creating new config.")
	config := configData{}

	fmt.Println("enter login credentials.")
	username := getInput("email")
	password := getInput("password")

	fmt.Println("enter details of project to diff.")
	p_id := getInput("project ID")

	config.Username = username
	config.Password = password
	config.P_ID = p_id
	config.unsavedChanges = true

	return config
}
