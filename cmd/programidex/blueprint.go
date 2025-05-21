package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Blueprint struct {
	Type        string   `json:"type"` // "app" or "module"
	WithModules bool     `json:"with_modules"`
	WithHugo    bool     `json:"with_hugo"`
	Directories []string `json:"directories"`
	GitHubRepo  string   `json:"github_repo"`
	GoModule    string   `json:"go_module"`
}

func loadBlueprint(configPath string) Blueprint {
	var blueprint Blueprint
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &blueprint)
	}
	return blueprint
}

func buildBlueprint(projectType, githubRepo, goModule string, reader *bufio.Reader) Blueprint {
	switch projectType {
	case "app":
		blueprint := Blueprint{
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
		return blueprint
	case "module":
		fmt.Print("Enter the module name: ")
		modName, _ := reader.ReadString('\n')
		modName = strings.TrimSpace(modName)
		return Blueprint{
			Type:        "module",
			WithHugo:    true,
			Directories: []string{modName + "/", "docs/", "hugo/"},
			GitHubRepo:  githubRepo,
			GoModule:    goModule,
		}
	default:
		fmt.Println("Invalid type. Please enter 'app' or 'module'.")
		os.Exit(1)
	}
	return Blueprint{}
}

func handleModuleBlueprintInstall(reader *bufio.Reader) {
	rootName := getCurrentDirName()
	defaultModuleName := rootName + "-module"
	modulesDir := filepath.Join("modules")
	defaultModuleDir := filepath.Join(modulesDir, defaultModuleName)
	moduleExists := fileExists(defaultModuleDir)
	var moduleName string

	if !moduleExists {
		fmt.Printf("No '%s' found in modules/.\n", defaultModuleName)
		fmt.Printf("Would you like to use the default module name '%s'? (y/n): ", defaultModuleName)
		resp, _ := reader.ReadString('\n')
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(resp)), "y") {
			moduleName = defaultModuleName
		} else {
			fmt.Print("Enter custom module name: ")
			moduleName, _ = reader.ReadString('\n')
			moduleName = strings.TrimSpace(moduleName)
		}
	} else {
		fmt.Printf("Default module '%s' already exists.\n", defaultModuleName)
		fmt.Print("Enter custom module name: ")
		moduleName, _ = reader.ReadString('\n')
		moduleName = strings.TrimSpace(moduleName)
	}

	// Show the blueprint
	fmt.Println("\nModule blueprint to be created:")
	fmt.Printf("modules/%s/\n", moduleName)
	fmt.Printf("modules/%s/main.go\n", moduleName)
	fmt.Printf("modules/%s/README.md\n", moduleName)

	fmt.Print("\nProceed with module creation? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		fmt.Println("Aborted module creation.")
		appendLog("programidex", "User aborted module blueprint creation.")
		return
	}

	err := installModuleBlueprint(moduleName)
	if err != nil {
		fmt.Printf("Failed to install module blueprint: %v\n", err)
		appendLog("programidex", fmt.Sprintf("Failed to install module blueprint: %v", err))
	} else {
		fmt.Printf("Module blueprint '%s' installed.\n", moduleName)
		appendLog("programidex", fmt.Sprintf("Module blueprint '%s' installed.", moduleName))
		updateConfigWithModule(moduleName)
	}
}

func installModuleBlueprint(moduleName string) error {
	moduleDir := filepath.Join("modules", moduleName)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return err
	}
	// Create a starter go file
	mainPath := filepath.Join(moduleDir, "main.go")
	if !fileExists(mainPath) {
		mainContent := `package main

import "fmt"

func main() {
    fmt.Println("Hello from ` + moduleName + `!")
}
`
		if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
			return err
		}
	}
	// Optionally, create a README or metadata file
	readmePath := filepath.Join(moduleDir, "README.md")
	if !fileExists(readmePath) {
		readmeContent := "# " + moduleName + "\n\nThis is a starter module for the DEX ecosystem."
		_ = os.WriteFile(readmePath, []byte(readmeContent), 0644)
	}
	return nil
}

func updateConfigWithModule(moduleName string) {
	configPath := filepath.Join(dexDir, programidexConfigFile)
	blueprint := loadBlueprint(configPath)
	moduleDir := filepath.Join("modules", moduleName)
	if !contains(blueprint.Directories, moduleDir) {
		blueprint.Directories = append(blueprint.Directories, moduleDir)
	}
	configBytes, _ := json.MarshalIndent(blueprint, "", "  ")
	_ = os.WriteFile(configPath, configBytes, 0644)
	appendLog("programidex", fmt.Sprintf("Config updated with new module: %s", moduleName))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
