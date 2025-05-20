package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
