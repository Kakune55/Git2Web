package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
)

type Config struct {
	RepoURL      string `json:"repo_url"`        // 仓库的URL
	TargetPath   string `json:"target_path"`     // 仓库克隆的目标路径
	WebhookPort  string `json:"webhook_port"`    // Webhook 监听的端口
	StaticPort   string `json:"static_port"`     // 静态文件服务监听的端口
	StaticPath   string `json:"static_path"`     // 静态内容服务的路径
	LogFilePath  string `json:"log_file_path"`   // 日志文件的路径
	LogMaxSizeMB int    `json:"log_max_size_mb"` // 最大日志文件大小（以MB为单位）
}

// loadConfig 从文件中读取配置，如果文件不存在则创建默认配置
func loadConfig(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 默认配置
		defaultConfig := Config{
			RepoURL:      "https://github.com/yourusername/yourrepo.git",
			TargetPath:   "./repo",
			WebhookPort:  "8081",
			StaticPort:   "8080",
			StaticPath:   "./repo",
			LogFilePath:  "./logs/server.log", // 日志文件路径
			LogMaxSizeMB: 5,                   // 日志文件最大大小 5MB
		}

		// 将默认配置写入文件
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

	// 读取配置文件
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

// customLogWriter 是一个自定义的日志写入器，监控日志文件大小并自动轮替
type customLogWriter struct {
	file        *os.File
	logDir      string
	logFile     string
	maxSize     int64
	currentSize int64
}

// newCustomLogWriter 创建一个新的 customLogWriter
func newCustomLogWriter(logFilePath string, maxSizeMB int) (*customLogWriter, error) {
	logDir := filepath.Dir(logFilePath)
	logFile := filepath.Base(logFilePath)
	maxSize := int64(maxSizeMB * 1024 * 1024) // 以字节为单位的最大日志大小

	// 打开或创建日志文件
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &customLogWriter{
		file:        file,
		logDir:      logDir,
		logFile:     logFile,
		maxSize:     maxSize,
		currentSize: stat.Size(),
	}, nil
}

// Write 实现 io.Writer 接口，写入日志并检查文件大小
func (w *customLogWriter) Write(p []byte) (n int, err error) {
	if w.currentSize+int64(len(p)) > w.maxSize {
		// 日志文件大小超过阈值，创建新日志文件
		if err := w.rotateLogFile(); err != nil {
			return 0, err
		}
	}
	n, err = w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

// rotateLogFile 轮替日志文件，创建新文件并重命名旧文件
func (w *customLogWriter) rotateLogFile() error {
	// 关闭当前日志文件
	if err := w.file.Close(); err != nil {
		return err
	}

	// 将当前日志文件重命名
	timestamp := time.Now().Format("20060102_150405")
	backupLogFile := filepath.Join(w.logDir, fmt.Sprintf("%s.%s", w.logFile, timestamp))
	if err := os.Rename(filepath.Join(w.logDir, w.logFile), backupLogFile); err != nil {
		return err
	}

	// 创建新的日志文件
	newFile, err := os.OpenFile(filepath.Join(w.logDir, w.logFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = newFile
	w.currentSize = 0
	return nil
}

func initLogging(config *Config) error {
	// 创建日志目录
	if err := os.Mkdir("./logs",0777); err != nil {
		return err
	}
	writer, err := newCustomLogWriter(config.LogFilePath, config.LogMaxSizeMB)
	if err != nil {
		return err
	}

	// 设置日志输出到文件
	log.SetOutput(io.MultiWriter(os.Stdout, writer)) // 日志同时输出到终端和文件
	return nil
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

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err.Error() == "already up-to-date" {
		log.Println("仓库已经是最新状态")
		return nil
	}
	return err
}

func webhookHandler(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("收到Webhook请求")
		err := pullRepo(config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("拉取仓库时出错: %v", err)
			return
		}
		fmt.Fprintln(w, "仓库成功更新")
	}
}

func serveStaticFiles(staticPath, port string) {
	fs := http.FileServer(http.Dir(staticPath))
	http.Handle("/", fs)
	log.Printf("静态文件服务器监听在端口 %s，提供 %s 的内容", port, staticPath)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("启动静态文件服务器时出错: %v", err)
	}
}

func serveWebhook(config *Config) {
	http.HandleFunc("/webhook", webhookHandler(config))
	log.Printf("Webhook 服务器监听在端口 %s URL 为 /webhook", config.WebhookPort)
	if err := http.ListenAndServe(":"+config.WebhookPort, nil); err != nil {
		log.Fatalf("启动 Webhook 服务器时出错: %v", err)
	}
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("加载配置时出错: %v", err)
	}

	// 初始化日志系统
	if err := initLogging(config); err != nil {
		log.Fatalf("初始化日志系统时出错: %v", err)
	}

	log.Println("载入配置")
	log.Println("同步自：", config.RepoURL)

	if _, err := os.Stat(config.TargetPath); os.IsNotExist(err) {
		log.Println("未找到仓库，正在克隆...")
		if err := cloneRepo(config); err != nil {
			log.Fatalf("克隆仓库时出错: %v", err)
		}
	}

	go serveStaticFiles(config.StaticPath, config.StaticPort)
	serveWebhook(config)
}
