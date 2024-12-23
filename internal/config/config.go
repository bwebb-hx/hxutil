package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	hx "github.com/bwebb-hx/hxutil/internal/hexaClient"
	"github.com/bwebb-hx/hxutil/internal/utils"
)

func configDir() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		utils.Fatal("failed to get home directory", err.Error())
	}
	path := filepath.Join(homePath, ".config", "hxutil")
	return path
}

func configFilePath() string {
	return filepath.Join(configDir(), "config.json")
}

func EnsureConfigDir() error {
	path := configDir()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0777)
	}
	return nil
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Project struct {
	P_ID          string `json:"p_id"`
	DisplayID     string `json:"display_id"`
	LastLoginUser string `json:"last_login_user"` // email used last for this project
}

type Config struct {
	LastLoginUser   string    `json:"last_login_user"`   // email used last time for hxutil
	LastUsedProject string    `json:"last_used_project"` // project used last time for hxutil
	Users           []User    `json:"users"`
	Projects        []Project `json:"projects"`
}

func (c *Config) AddProject() *Project {
	// get pid from user
	p_id := utils.GetInput("Project ID")

	// determine the user credentials to login with
	c.SelectUserAndLogin("")

	// get project details
	bytes, err := hx.GetApi(hx.UN_GetProjectSettingsAPI.URI, map[string]string{"p_id": p_id})
	if err != nil {
		utils.Fatal("failed to get project details", err.Error())
	}
	var resp hx.UN_GetProjectSettingsResponse
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		utils.Fatal("failed to unmarshal project details response", err.Error())
	}

	project := Project{
		P_ID:      p_id,
		DisplayID: resp.DisplayID,
	}
	c.Projects = append(c.Projects, project)
	c.Save()

	return &project
}

func (c *Config) SelectProject() *Project {
	if len(c.Projects) > 0 {
		fmt.Println("Existing projects:")
		for i, project := range c.Projects {
			fmt.Printf("%v) %s [%s]\n", i+1, project.DisplayID, project.P_ID)
		}
		input := utils.GetInput("Choose a project (or \"new\")")

		if input == "new" {
			return c.AddProject()
		}
		index, err := strconv.Atoi(input)
		if err != nil {
			utils.Fatal("failed to parse index number", err.Error())
		}
		for i, project := range c.Projects {
			if i+1 == index {
				return &project
			}
		}
	}

	return c.AddProject()
}

func (c *Config) SelectUserAndLogin(p_id string) {
	// determine if a "last login user" is applicable
	lastLoginUser := ""
	if p_id != "" {
		// when p_id is passed, look in the corresponding project
		for _, project := range c.Projects {
			if project.P_ID == p_id {
				lastLoginUser = project.LastLoginUser
				break
			}
		}
		if lastLoginUser == "" {
			utils.Hint("(No login user found for given project)")
		} else {
			fmt.Println("Last login user for this project:", lastLoginUser)
		}
	} else {
		// if no p_id is passed, check the last logged in user overall
		if c.LastLoginUser != "" {
			lastLoginUser = c.LastLoginUser
			fmt.Println("Last login user:", lastLoginUser)
		}
	}

	// if a last login user is found, try to use that
	if lastLoginUser != "" {
		if utils.YesOrNo("Login with this user?") {
			password := ""
			for _, user := range c.Users {
				if user.Email == lastLoginUser {
					password = user.Password
					break
				}
			}
			if password == "" {
				utils.Error("failed to find registered user in config", "")
			} else {
				hx.Login(lastLoginUser, password)
				return
			}
		}
	}

	// choose an existing user or register a new one
	if len(c.Users) == 0 {
		utils.Hint("(no existing users found)")
		user := c.AddNewUser()
		if p_id != "" {
			c.SetProjectLastUser(p_id, user.Email)
		}
		return
	}
	for i, user := range c.Users {
		fmt.Println(i+1, ")", user.Email)
	}
	input := utils.GetInput("Choose user (or \"new\")")
	if strings.ToLower(input) == "new" {
		user := c.AddNewUser()
		if p_id != "" {
			c.SetProjectLastUser(p_id, user.Email)
		}
		return
	}

	// find the corresponding user
	index, err := strconv.Atoi(input)
	if err != nil {
		utils.Error("failed to parse input", err.Error())
	}
	for i, user := range c.Users {
		if i+1 == index {
			hx.Login(user.Email, user.Password)
			if p_id != "" {
				c.SetProjectLastUser(p_id, user.Email)
			}
			return
		}
	}
	utils.Error("failed to login", "entered index invalid")
}

func (c *Config) SetProjectLastUser(p_id, userEmail string) {
	for i, project := range c.Projects {
		if project.P_ID == p_id {
			c.Projects[i].LastLoginUser = userEmail
			c.Save()
			return
		}
	}
	utils.Error("failed to set last login user for project", "matching p_id not found")
}

func (c *Config) AddNewUser() *User {
	email := utils.GetInput("User email")
	password := utils.GetInput("Password")

	// attempt login
	hx.Login(email, password)

	// add to config
	user := User{
		Email:    email,
		Password: password,
	}
	c.Users = append(c.Users, user)
	c.LastLoginUser = email
	c.Save()

	return &user
}

func (c Config) Save() {
	path := configFilePath()

	bytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		utils.Error("failed to save config", err.Error())
		return
	}

	err = os.WriteFile(path, bytes, 0777)
	if err != nil {
		utils.Error("failed to save config", err.Error())
	}
}

func GetConfig() *Config {
	if err := EnsureConfigDir(); err != nil {
		utils.Fatal("error while ensuring config directory", err.Error())
	}
	path := configFilePath()

	// if config doesn't exist yet, return an empty struct
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &Config{}
	}
	if err != nil {
		utils.Fatal("error while checking for config file", err.Error())
	}

	// read the existing config file
	configBytes, err := os.ReadFile(path)
	if err != nil {
		utils.Error("failed to read config", err.Error())
		return nil
	}

	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		utils.Error("failed to unmarshal config data", err.Error())
		return nil
	}
	return &config
}
