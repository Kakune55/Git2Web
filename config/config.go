package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// AppVersion 应用版本
const AppVersion = "1.3.0"

// Config 应用配置
type Config struct {
	RepoURL         string   `json:"repo_url"`
	UpdateOnStart   bool     `json:"update_on_start"`
	TargetPathA     string   `json:"target_path_a"`
	TargetPathB     string   `json:"target_path_b"`
	ActivePartition string   `json:"active_partition"`
	WebhookPort     string   `json:"webhook_port"`
	WebhookSecret   string   `json:"webhook_secret"`
	StaticPort      string   `json:"static_port"`
	StaticPath      string   `json:"static_path"`
	LogFilePath     string   `json:"log_file_path"`
	LogMaxSizeMB    int      `json:"log_max_size_mb"`
	RepoAuth        RepoAuth `json:"repo_auth"`
	LfsEnabled      bool     `json:"lfs_enabled"`
	Version         string   `json:"version"`
}

// RepoAuth 仓库认证信息
type RepoAuth struct {
	Enabled  bool   `json:"enabled"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// getEnv 获取环境变量，若不存在则返回默认值
func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// getEnvBool 获取布尔类型环境变量
func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

// getEnvInt 获取整型环境变量
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

// LoadConfig 从文件加载配置，若文件不存在则优先用环境变量初始化
func LoadConfig(filename string) (*Config, error) {
	// 创建目录

	configDir, _ := filepath.Split(filename)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, err
		}
	}

	// 如果配置文件不存在，则创建默认配置（优先读取环境变量）
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		defaultConfig := Config{
			RepoURL:         getEnv("REPO_URL", "https://github.com/yourusername/yourrepo.git"),
			UpdateOnStart:   getEnvBool("UPDATE_ON_START", true),
			TargetPathA:     getEnv("TARGET_PATH_A", "./data/repo_a"),
			TargetPathB:     getEnv("TARGET_PATH_B", "./data/repo_b"),
			ActivePartition: getEnv("ACTIVE_PARTITION", "a"),
			WebhookPort:     getEnv("WEBHOOK_PORT", "8081"),
			WebhookSecret:   getEnv("WEBHOOK_SECRET", ""),
			StaticPort:      getEnv("STATIC_PORT", "8080"),
			StaticPath:      getEnv("STATIC_PATH", "./data/repo"),
			LogFilePath:     getEnv("LOG_FILE_PATH", "./logs/server.log"),
			LogMaxSizeMB:    getEnvInt("LOG_MAX_SIZE_MB", 5),
			RepoAuth: RepoAuth{
				Enabled:  getEnvBool("REPO_AUTH_ENABLED", false),
				Email:    getEnv("REPO_AUTH_EMAIL", "example@example.com"),
				Password: getEnv("REPO_AUTH_PASSWORD", "1234"),
			},
			LfsEnabled: getEnvBool("LFS_ENABLED", false),
			Version:    AppVersion,
		}

		configData, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filename, configData, 0644); err != nil {
			return nil, err
		}
		log.Printf("配置文件已创建: %s", filename)
		log.Println("请检查并编辑配置文件后重新运行以生效更改。")
	}

	file, err := os.ReadFile(filename)
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

// 确保目录存在
func ensureDirExists(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetActiveTargetPath 获取当前活动分区的路径
func (c *Config) GetActiveTargetPath() string {
	if c.ActivePartition == "b" {
		return c.TargetPathB
	}
	return c.TargetPathA
}

// GetInactiveTargetPath 获取非活动分区的路径
func (c *Config) GetInactiveTargetPath() string {
	if c.ActivePartition == "b" {
		return c.TargetPathA
	}
	return c.TargetPathB
}

// SwitchActivePartition 切换活动分区
func (c *Config) SwitchActivePartition() {
	if c.ActivePartition == "a" {
		c.ActivePartition = "b"
	} else {
		c.ActivePartition = "a"
	}
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig(configPath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
