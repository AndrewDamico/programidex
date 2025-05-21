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

const dexDir = ".dex"
const programidexConfigFile = ".programidex.json"
const portfolidexConfigFile = ".portfolidex.json"
const dexLogFile = "dex.log"

func main() {
	reader := bufio.NewReader(os.Stdin)
	os.MkdirAll(dexDir, 0755)
	configPath := filepath.Join(dexDir, programidexConfigFile)

	var blueprint Blueprint
	var githubRepo string

	// Load existing config if present
	if fileExists(configPath) {
		blueprint = loadBlueprint(configPath)
		for {
			fmt.Println("\nPortfoliDEX Project Detected.")
			fmt.Println("1. Install a new module into the DEX")
			fmt.Println("2. View current config")
			fmt.Println("3. Exit")
			fmt.Println("4. Install a Hugo site") // <-- new option
			fmt.Print("Choose an option [1-4]: ")
			choice, _ := reader.ReadString('\n')
			choice = strings.TrimSpace(choice)
			switch choice {
			case "1":
				handleModuleBlueprintInstall(reader)
				// Optionally reload config if you want to show updated config after install
				blueprint = loadBlueprint(configPath)
			case "2":
				configBytes, _ := json.MarshalIndent(blueprint, "", "  ")
				fmt.Println("\nCurrent config:")
				fmt.Println(string(configBytes))
			case "3":
				fmt.Println("Exiting.")
				return
			case "4":
				handleHugoSiteInstall(reader) // <-- new handler
			default:
				fmt.Println("Invalid choice.")
			}
		}
		return
	}

	fmt.Println("Initialize a new Go project with programidex.")
	projectType := promptProjectType(reader)
	githubRepo = setupGitHub(reader)
	goModule := promptGoModule(reader, projectType)

	blueprint = buildBlueprint(projectType, githubRepo, goModule, reader)
	blueprint.Directories = append(blueprint.Directories, ".programidex/")
	blueprint.Directories = append(blueprint.Directories, ".dex/")

	// --- New: Ask to install a module blueprint if app supports modules ---
	if blueprint.Type == "app" && blueprint.WithModules {
		fmt.Print("\nWould you like to install a module blueprint now? (y/n): ")
		resp, _ := reader.ReadString('\n')
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(resp)), "y") {
			handleModuleBlueprintInstall(reader)
		}
	}

	configBytes, _ := json.MarshalIndent(blueprint, "", "  ")
	fmt.Println("\nProposed configuration:")
	fmt.Println(string(configBytes))
	fmt.Print("\nProceed with directory creation and Go initialization? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		fmt.Println("Aborted.")
		appendLog("programidex", "User aborted initialization.")
		return
	}

	createDirectories(blueprint.Directories)
	initializeGoModule(blueprint.GoModule)
	ensureMainGoForApp(&blueprint)
	_ = os.WriteFile(configPath, configBytes, 0644)
	appendLog("programidex", "Saved configuration.")

	fmt.Println("\nInitialization complete. Config and log saved in .programidex/")
}

// --- Utility and workflow functions ---

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func appendLog(appName, msg string) {
	logPath := filepath.Join(dexDir, dexLogFile)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		timestamp := time.Now().Format(time.RFC3339)
		f.WriteString(fmt.Sprintf("[%s][%s] %s\n", appName, timestamp, msg))
	}
}

func handleExistingProject(blueprint *Blueprint, reader *bufio.Reader, configPath string) {
	fmt.Println("This project has already been initialized by programidex.")
	if blueprint.GitHubRepo == "" {
		fmt.Println("GitHub namespace/repo is missing.")
		fmt.Println("Would you like to set up GitHub now? (y/n): ")
		resp, _ := reader.ReadString('\n')
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(resp)), "y") {
			blueprint.GitHubRepo = setupGitHub(reader)
			configBytes, _ := json.MarshalIndent(blueprint, "", "  ")
			_ = os.WriteFile(configPath, configBytes, 0644)
			appendLog("programidex", "GitHub repo set after re-initialization.")
			fmt.Println("GitHub repo set.")
		} else {
			fmt.Println("You can set up GitHub later.")
			appendLog("programidex", "Skipped GitHub setup on re-initialization.")
		}
	} else {
		fmt.Println("Nothing to update in config, but checking for missing directories or files...")
		appendLog("programidex", "Attempted re-initialization; checking for missing directories or files.")
	}
	ensureMainGoForApp(blueprint)
	fmt.Println("Repair (if needed) complete. Exiting.")
}

func promptProjectType(reader *bufio.Reader) string {
	fmt.Print("Are you initializing an 'app' or a 'module'? ")
	projectType, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(projectType))
}

func promptGoModule(reader *bufio.Reader, projectType string) string {
	if projectType == "app" || projectType == "module" {
		// Try to auto-detect from git remote
		defaultModule := ""
		remoteURL, err := getGitRemoteURL()
		if err == nil && remoteURL != "" {
			remoteURL = strings.TrimSuffix(remoteURL, ".git")
			if strings.HasPrefix(remoteURL, "https://") {
				defaultModule = strings.TrimPrefix(remoteURL, "https://")
			} else if strings.HasPrefix(remoteURL, "git@") {
				defaultModule = strings.TrimPrefix(remoteURL, "git@")
				defaultModule = strings.Replace(defaultModule, ":", "/", 1)
			} else {
				defaultModule = remoteURL
			}
		}
		fmt.Printf("Enter the Go application path (e.g., github.com/youruser/yourrepo) [%s]: ", defaultModule)
		goModule, _ := reader.ReadString('\n')
		goModule = strings.TrimSpace(goModule)
		if goModule == "" {
			goModule = defaultModule
		}
		return goModule
	}
	return ""
}

func createDirectories(dirs []string) {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Failed to create %s: %v\n", dir, err)
			appendLog("programidex", fmt.Sprintf("Failed to create %s: %v", dir, err))
		} else {
			fmt.Printf("Created %s\n", dir)
			appendLog("programidex", fmt.Sprintf("Created %s", dir))
		}
	}
}

func initializeGoModule(goModule string) {
	if !fileExists("go.mod") && goModule != "" {
		cmd := exec.Command("go", "mod", "init", goModule)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to initialize go.mod: %v\n", err)
			appendLog("programidex", fmt.Sprintf("Failed to initialize go.mod: %v", err))
		} else {
			fmt.Println("Initialized go.mod")
			appendLog("programidex", "Initialized go.mod")
		}
	}
}

func ensureMainGoForApp(blueprint *Blueprint) {
	if blueprint.Type != "app" {
		return
	}
	folderName := getCurrentDirName()
	if blueprint.GoModule != "" {
		parts := strings.Split(blueprint.GoModule, "/")
		folderName = parts[len(parts)-1]
	}
	mainDir := filepath.Join("cmd", folderName)
	mainPath := filepath.Join(mainDir, "main.go")
	needsMain := false
	if !fileExists(mainDir) {
		if err := os.MkdirAll(mainDir, 0755); err == nil {
			fmt.Printf("Created missing directory: %s\n", mainDir)
			appendLog("programidex", fmt.Sprintf("Created missing directory: %s", mainDir))
			needsMain = true
		}
	}
	if !fileExists(mainPath) {
		needsMain = true
	}
	if needsMain {
		mainContent := `package main

import "fmt"

func main() {
    fmt.Println("Welcome to ` + folderName + `!")
}
`
		_ = os.WriteFile(mainPath, []byte(mainContent), 0644)
		fmt.Printf("Created starter Go file: %s\n", mainPath)
		appendLog("programidex", fmt.Sprintf("Created starter Go file: %s", mainPath))
	}
}
