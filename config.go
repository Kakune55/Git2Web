package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	RepoURL      string `json:"repo_url"`
	TargetPath   string `json:"target_path"`
	WebhookPort  string `json:"webhook_port"`
	StaticPort   string `json:"static_port"`
	StaticPath   string `json:"static_path"`
	LogFilePath  string `json:"log_file_path"`
	LogMaxSizeMB int    `json:"log_max_size_mb"`
	RepoAuth     map[string]string `json:"repo_auth"`
}

func loadConfig(filename string) (*Config, error) {
	// 如果配置文件不存在，则创建默认配置
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		defaultConfig := Config{
			RepoURL:      "https://github.com/yourusername/yourrepo.git",
			TargetPath:   "./repo",
			WebhookPort:  "8081",
			StaticPort:   "8080",
			StaticPath:   "./repo",
			LogFilePath:  "./logs/server.log",
			LogMaxSizeMB: 5,
			RepoAuth: map[string]string{"enabled": "false","username": "yourusername","password": "yourpassword"},
		}

		configData, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filename, configData, 0644); err != nil {
			return nil, err
		}
		log.Printf("配置文件已创建: %s", filename)
		log.Println("程序已退出，请编辑配置文件后重新运行。")
		os.Exit(0)
	}

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
