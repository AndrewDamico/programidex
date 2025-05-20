package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Blueprint struct {
	Type        string   `json:"type"` // "app" or "module"
	WithModules bool     `json:"with_modules"`
	WithHugo    bool     `json:"with_hugo"`
	Directories []string `json:"directories"`
	GitHubRepo  string   `json:"github_repo"`
	GoModule    string   `json:"go_module"`
}

const programidexDir = ".programidex"
const configFile = "init_config.json"
const logFile = "init.log"

func getCurrentDirName() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(wd)
}

func getGitRemoteURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func appendLog(msg string) {
	logPath := filepath.Join(programidexDir, logFile)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		timestamp := time.Now().Format(time.RFC3339)
		f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	os.MkdirAll(programidexDir, 0755)
	configPath := filepath.Join(programidexDir, configFile)

	var blueprint Blueprint
	var githubRepo string

	// Check for existing config
	if fileExists(configPath) {
		data, err := os.ReadFile(configPath)
		if err == nil {
			json.Unmarshal(data, &blueprint)
		}
		fmt.Println("This project has already been initialized by programidex.")
		if blueprint.GitHubRepo == "" {
			fmt.Println("GitHub namespace/repo is missing.")
			fmt.Println("Would you like to set up GitHub now? (y/n): ")
			resp, _ := reader.ReadString('\n')
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(resp)), "y") {
				// Move GitHub setup logic here
				githubRepo = setupGitHub(reader)
				blueprint.GitHubRepo = githubRepo
				// Save updated config and log
				configBytes, _ := json.MarshalIndent(blueprint, "", "  ")
				_ = os.WriteFile(configPath, configBytes, 0644)
				appendLog("GitHub repo set after re-initialization.")
				fmt.Println("GitHub repo set. Exiting.")
				return
			} else {
				fmt.Println("You can set up GitHub later. Exiting.")
				appendLog("Skipped GitHub setup on re-initialization.")
				return
			}
		} else {
			fmt.Println("Nothing to update. Exiting.")
			appendLog("Attempted re-initialization; nothing to update.")
			return
		}
	}

	fmt.Println("Initialize a new Go project with programidex.")
	fmt.Print("Are you initializing an 'app' or a 'module'? ")
	projectType, _ := reader.ReadString('\n')
	projectType = strings.TrimSpace(strings.ToLower(projectType))

	githubRepo = setupGitHub(reader)

	var goModule string
	if projectType == "app" || projectType == "module" {
		fmt.Print("Enter the Go module path (e.g., github.com/youruser/yourrepo): ")
		goModule, _ = reader.ReadString('\n')
		goModule = strings.TrimSpace(goModule)
	}

	switch projectType {
	case "app":
		blueprint = Blueprint{
			Type:        "app",
			WithHugo:    true,
			Directories: []string{"cmd/", "internal/", "docs/", "hugo/"},
			GitHubRepo:  githubRepo,
			GoModule:    goModule,
		}
		fmt.Print("Will this app have modules? (y/n): ")
		modulesResp, _ := reader.ReadString('\n')
		blueprint.WithModules = strings.HasPrefix(strings.ToLower(strings.TrimSpace(modulesResp)), "y")
		if blueprint.WithModules {
			blueprint.Directories = append(blueprint.Directories, "modules/")
		}
	case "module":
		fmt.Print("Enter the module name: ")
		modName, _ := reader.ReadString('\n')
		modName = strings.TrimSpace(modName)
		blueprint = Blueprint{
			Type:        "module",
			WithHugo:    true,
			Directories: []string{modName + "/", "docs/", "hugo/"},
			GitHubRepo:  githubRepo,
			GoModule:    goModule,
		}
	default:
		fmt.Println("Invalid type. Please enter 'app' or 'module'.")
		appendLog("Invalid project type entered.")
		return
	}

	// Always add .programidex for logs/config
	blueprint.Directories = append(blueprint.Directories, ".programidex/")

	// Display config
	configBytes, _ := json.MarshalIndent(blueprint, "", "  ")
	fmt.Println("\nProposed configuration:")
	fmt.Println(string(configBytes))
	fmt.Print("\nProceed with directory creation and Go initialization? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		fmt.Println("Aborted.")
		appendLog("User aborted initialization.")
		return
	}

	// Create directories
	for _, dir := range blueprint.Directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Failed to create %s: %v\n", dir, err)
			appendLog(fmt.Sprintf("Failed to create %s: %v", dir, err))
		} else {
			fmt.Printf("Created %s\n", dir)
			appendLog(fmt.Sprintf("Created %s", dir))
		}
	}

	// Initialize Go module if not already done
	if !fileExists("go.mod") && blueprint.GoModule != "" {
		cmd := exec.Command("go", "mod", "init", blueprint.GoModule)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to initialize go.mod: %v\n", err)
			appendLog(fmt.Sprintf("Failed to initialize go.mod: %v", err))
		} else {
			fmt.Println("Initialized go.mod")
			appendLog("Initialized go.mod")
		}
	}

	// Save config and log
	_ = os.WriteFile(configPath, configBytes, 0644)
	appendLog("Saved configuration.")

	fmt.Println("\nInitialization complete. Config and log saved in .programidex/")
}

// Helper function for GitHub setup
func setupGitHub(reader *bufio.Reader) string {
	remoteURL, err := getGitRemoteURL()
	if err == nil && remoteURL != "" {
		fmt.Printf("Detected existing GitHub remote: %s\n", remoteURL)
		return remoteURL
	}
	fmt.Println("No GitHub remote detected.")
	fmt.Println("1. Set up GitHub repo for me")
	fmt.Println("2. Enter repo name manually")
	fmt.Print("Choose an option [1/2]: ")
	opt, _ := reader.ReadString('\n')
	opt = strings.TrimSpace(opt)
	if opt == "1" || opt == "" {
		repoName := getCurrentDirName()
		fmt.Printf("Setting up GitHub repo named '%s'...\n", repoName)
		return repoName
	} else {
		fmt.Print("Enter the GitHub repo name or URL: ")
		githubRepo, _ := reader.ReadString('\n')
		return strings.TrimSpace(githubRepo)
	}
}
