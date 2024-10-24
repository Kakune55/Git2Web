package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	git "github.com/go-git/go-git/v5"
)

type Config struct {
    RepoURL   string `json:"repo_url"`
    TargetPath string `json:"target_path"`
    Port      string `json:"port"`
}

func loadConfig(filename string) (*Config, error) {
    // Check if the configuration file exists
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        // Create default configuration
        defaultConfig := Config{
            RepoURL:   "https://github.com/yourusername/yourrepo.git",
            TargetPath: "./yourrepo",
            Port:      "8080",
        }
        // Write default configuration to file
        configData, err := json.MarshalIndent(defaultConfig, "", "  ")
        if err != nil {
            return nil, err
        }
        if err := os.WriteFile(filename, configData, 0644); err != nil {
            return nil, err
        }
        log.Printf("Configuration file does not exist, created default configuration file: %s", filename)
        return &defaultConfig, nil
    }

    // Read configuration file
    file, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := json.Unmarshal(file, &config); err != nil {
        return nil, err
    }
    return &config, nil
}

func cloneRepo(config *Config) error {
    _, err := git.PlainClone(config.TargetPath, false, &git.CloneOptions{
        URL: config.RepoURL,
    })
    return err
}

func pullRepo(config *Config) error {
    repo, err := git.PlainOpen(config.TargetPath)
    if err != nil {
        return err
    }

    w, err := repo.Worktree()
    if err != nil {
        return err
    }

    // Pull latest changes
    err = w.Pull(&git.PullOptions{RemoteName: "origin"})
    return err
}

func webhookHandler(config *Config) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("Received Webhook request")
        err := pullRepo(config)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf("Error pulling repository: %v", err)
            return
        }
        fmt.Fprintln(w, "Repository updated successfully")
    }
}

func main() {
    config, err := loadConfig("config.json")
    if err != nil {
        log.Fatalf("Error loading configuration: %v", err)
    }

    // Check if the repository exists
    if _, err := os.Stat(config.TargetPath); os.IsNotExist(err) {
        log.Println("Repository not found, cloning...")
        if err := cloneRepo(config); err != nil {
            log.Fatalf("Error cloning repository: %v", err)
        }
    }

    // Start Webhook server
    http.HandleFunc("/webhook", webhookHandler(config))
    log.Printf("Listening on port %s...", config.Port)
    if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
}