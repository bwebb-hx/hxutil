package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const OUT = ".as_temp"

func main() {
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

	// login to hexabase
	if err := login(); err != nil {
		log.Fatal("error occurred while logging in:", err)
	}

	// download all action scripts
	if err := downloadActionScripts(); err != nil {
		log.Fatal("error occurred while downloading action scripts:", err)
	}

	// compare all files for the given datastore
	compareAllFiles(absPath)
}

func compareAllFiles(absPath string) {
	// get the relative path to the scripts
	pathToTemp := filepath.Join(absPath, OUT)
	files, err := os.ReadDir(pathToTemp)
	if err != nil {
		log.Fatal(err)
	}
	datastore := getInput("Datastore name")
	pathToDatastoreDir := filepath.Join(pathToTemp, files[0].Name(), datastore)

	// for all files, diff with the corresponding file found in the current directory or below
	actionScriptFiles, err := os.ReadDir(pathToDatastoreDir)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal("failed to find file:", err)
		}
	}

	fmt.Println("== SUMMARY ==")
	fmt.Println("total comps:", totalComps)
	fmt.Println("Files that don't match remote:")
	for _, file := range diffFiles {
		fmt.Println(file)
	}

	if err := os.RemoveAll(pathToTemp); err != nil {
		fmt.Println("failed to remove temp dir:", err)
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
			fmt.Println("local:", local)
			fmt.Println("remote:", remote)
			fmt.Println("~~~~~~~~~~~~~~~~")
			fmt.Println(string(output))
			fmt.Println("~~~~~~~~~~~~~~~~")
			return true
		}
		log.Println("failed to diff:", err)
	}
	return false
}

func login() error {
	user := getInput("username")
	pass := getInput("password")
	cmd := exec.Command("hx", "login", "--email="+user, "--password="+pass)
	return cmd.Run()
}

// hx actions:scripts:download_all [PROJECT_ID]
func downloadActionScripts() error {
	p_id := getInput("Project ID")
	cmd := exec.Command("hx", "actions:scripts:download_all", p_id, "--output="+OUT)
	return cmd.Run()
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
