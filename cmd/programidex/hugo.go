package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func handleHugoSiteInstall(reader *bufio.Reader) {
	fmt.Print("Enter the directory name for the Hugo site [hugo]: ")
	dir, _ := reader.ReadString('\n')
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = "hugo"
	}

	projectRoot := findProjectRoot()
	if projectRoot == "" {
		fmt.Println("Could not find project root (missing .dex directory). Aborting.")
		appendLog("programidex", "Hugo site creation aborted: project root not found.")
		return
	}

	absDir := filepath.Join(projectRoot, dir)

	if fileExists(absDir) {
		entries, err := os.ReadDir(absDir)
		if err != nil {
			fmt.Printf("Could not read directory '%s'. Aborting.\n", absDir)
			appendLog("programidex", fmt.Sprintf("Hugo site creation aborted: could not read %s.", absDir))
			return
		}
		if len(entries) > 0 {
			fmt.Printf("Directory '%s' is not empty. Aborting Hugo site creation.\n", absDir)
			appendLog("programidex", fmt.Sprintf("Hugo site creation aborted: %s not empty.", absDir))
			return
		}
	}

	fmt.Printf("Creating Hugo site in '%s'...\n", absDir)
	cmd := exec.Command("hugo", "new", "site", absDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to create Hugo site: %v\n", err)
		appendLog("programidex", fmt.Sprintf("Failed to create Hugo site: %v", err))
		return
	}

	fmt.Printf("Hugo site created in '%s'.\n", absDir)
	appendLog("programidex", fmt.Sprintf("Hugo site created in '%s'.", absDir))
}

// Helper to find the project root by searching for .dex upwards
func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if fileExists(filepath.Join(dir, ".dex")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
