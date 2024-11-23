package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const OUT = ".as_temp"
const CONFIG = ".hx-tools.json"

type configData struct {
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	P_ID           string   `json:"p_id"`
	Datastores     []string `json:"datastores"`
	unsavedChanges bool
	LastLogin      time.Time `json:"last_login"`
}

func main1() {
	// first, get the current directory where we are in the terminal
	if len(os.Args) < 2 {
		fmt.Println("Usage: as_checker <path>")
	}
	path := os.Args[1]

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Error resolving path: %d\n", err)
		return
	}

	// load config, if it exists
	config := loadConfig()

	// login to hexabase
	if time.Since(config.LastLogin) > time.Hour {
		err = login(&config)
		if err != nil {
			log.Fatal("error occurred while logging in:", err)
		}
		config.LastLogin = time.Now()
	}

	// download all action scripts
	if config.P_ID == "" {
		config.P_ID = getInput("Project ID")
		config.unsavedChanges = true
	}
	if err := downloadActionScripts(config.P_ID, len(config.Datastores) == 0); err != nil {
		log.Fatal("error occurred while downloading action scripts:", err)
	}

	// compare all files for the given datastore
	if len(config.Datastores) == 0 {
		datastoresInput := getInput("Datastore names (use comma as delim)")
		config.Datastores = make([]string, 0)
		for _, datastore := range strings.Split(datastoresInput, ",") {
			datastore = strings.TrimSpace(datastore)
			config.Datastores = append(config.Datastores, datastore)
		}
		config.unsavedChanges = true
	}
	compareAllFiles(absPath, config.Datastores)

	if config.unsavedChanges && strings.ToLower(getInput("Save project/user details for next time? [Y/n]")) == "y" {
		saveConfig(config)
	}
}

func compareAllFiles(absPath string, datastores []string) {
	// get the relative path to the scripts
	pathToTemp := filepath.Join(absPath, OUT)
	files, err := os.ReadDir(pathToTemp)
	if err != nil {
		log.Fatal(err)
	}

	for _, datastore := range datastores {
		pathToDatastoreDir := filepath.Join(pathToTemp, files[0].Name(), datastore)
		fmt.Println("\nChecking ActionScripts for:", datastore)
		compareFilesInDatastore(pathToDatastoreDir, absPath)
	}

	if err := os.RemoveAll(pathToTemp); err != nil {
		fmt.Println("failed to remove temp dir:", err)
	}
}

func compareFilesInDatastore(pathToDatastoreDir, absPath string) {
	// for all files, diff with the corresponding file found in the current directory or below
	actionScriptFiles, err := os.ReadDir(pathToDatastoreDir)
	if err != nil {
		fmt.Printf("Failed to open directory %s: %d\n", pathToDatastoreDir, err)
		return
	}

	diffFiles := make([]string, 0)
	totalComps := 0

	for _, script := range actionScriptFiles {
		searchPieces := strings.Split(script.Name(), "-") // file names have "-" in the middle, which divides script name and pre/post.js

		err := filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasPrefix(d.Name(), searchPieces[0]) {
				if strings.HasSuffix(d.Name(), searchPieces[1]) {
					// match found! get diff results
					if diff(path, filepath.Join(pathToDatastoreDir, script.Name()), script.Name()) {
						diffFiles = append(diffFiles, d.Name())
					}
					totalComps++
				}
			}
			return nil
		})
		if err != nil {
			log.Println("failed to find file:", err)
		}
	}

	fmt.Println("== SUMMARY ==")
	fmt.Println("total comps:", totalComps)
	if len(diffFiles) == 0 {
		fmt.Println("No differences found ✔")
	} else {
		fmt.Println("Files that don't match remote:")
		for _, file := range diffFiles {
			fmt.Println(file)
		}
	}
}

// returns true if a difference is found
func diff(local, remote, fileName string) bool {
	cmd := exec.Command("diff", "-w", local, remote)
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 {
			// differences found
			fmt.Println("Differences found for", fileName, "...")
			fmt.Println("~~~~~~~~~~~~~~~~")
			fmt.Println(string(output))
			fmt.Println("~~~~~~~~~~~~~~~~")
			return true
		}
		log.Println("failed to diff:", err)
	}
	return false
}

func login(config *configData) error {
	if config.Username == "" {
		config.Username = getInput("username")
		config.unsavedChanges = true
	}
	if config.Password == "" {
		config.Password = getInput("password")
	}
	cmd := exec.Command("hx", "login", "--email="+config.Username, "--password="+config.Password)
	return cmd.Run()
}

// hx actions:scripts:download_all [PROJECT_ID]
func downloadActionScripts(p_id string, showDatastores bool) error {
	fmt.Println("\nLoading ActionScripts (this may take a second)")
	cmd := exec.Command("hx", "actions:scripts:download_all", p_id, "--output="+OUT)
	if err := cmd.Run(); err != nil {
		return err
	}

	// show all projects found to the user
	if !showDatastores {
		return nil
	}
	files, err := os.ReadDir(OUT)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("failed to find project folder")
	}

	projectFolder := files[0] // we expect there to be one folder in the output folder; the project folder
	if !projectFolder.IsDir() {
		return errors.New("no project folder?")
	}
	datastoresFolderPath := filepath.Join(OUT, projectFolder.Name())
	files, err = os.ReadDir(datastoresFolderPath)
	if err != nil {
		return err
	}

	fmt.Println("Datastores Found:")
	for _, file := range files {
		fmt.Print(file.Name() + ", ")
	}
	fmt.Println()
	fmt.Println()
	return nil
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
		return configData{}
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
	fmt.Println("Datastores:", strings.Join(config.Datastores, ", "))
	if strings.ToLower(getInput("\n\nUse above config? [y/n]")) != "y" {
		fmt.Println("ignoring existing config.")
		return configData{}
	}

	return config
}
