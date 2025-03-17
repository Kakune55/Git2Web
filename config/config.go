package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type RepoAuthConfig struct {
	Enabled		bool   `json:"enabled"`
	Email		string `json:"email"`
	Password	string `json:"password"`
}

type Config struct {
	RepoURL        string        `json:"repo_url"`
	TargetPath     string        `json:"target_path"`
	WebhookPort    string        `json:"webhook_port"`
	WebhookSecret  string        `json:"webhook_secret"`
	StaticPort     string        `json:"static_port"`
	StaticPath     string        `json:"static_path"`
	LogFilePath    string        `json:"log_file_path"`
	LogMaxSizeMB   int           `json:"log_max_size_mb"`
	RepoAuth       RepoAuthConfig `json:"repo_auth"`
	LfsEnabled     bool          `json:"lfs_enabled"`
	Version        string        `json:"version"`
}

// 应用版本号
const AppVersion = "1.2.0"

func LoadConfig(filename string) (*Config, error) {
	// 创建conf目录
	configDir := "etc"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, err
		}
	}
	
	configPath := filepath.Join(configDir, filename)
	
	// 如果配置文件不存在，则创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := Config{
			RepoURL:       "https://github.com/yourusername/yourrepo.git",
			TargetPath:    "./data/repo",
			WebhookPort:   "8081",
			WebhookSecret: "",  // 默认为空，不启用验证
			StaticPort:    "8080",
			StaticPath:    "./data/repo",
			LogFilePath:   "./logs/server.log",
			LogMaxSizeMB:  5,
			RepoAuth: RepoAuthConfig{
				Enabled:   false,
				Email:     "example@example.com",
				Password:  "1234",
			},
			LfsEnabled:    false,
			Version:       AppVersion,
		}

		configData, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			return nil, err
		}
		log.Printf("配置文件已创建: %s", configPath)
		log.Println("程序已退出，请编辑配置文件后重新运行。")
		os.Exit(0)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	
	// 确保版本信息是最新的
	config.Version = AppVersion
	
	return &config, nil
}
