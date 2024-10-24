package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	git "github.com/go-git/go-git/v5"
)

// Config 结构体用于存储仓库和服务器设置
type Config struct {
	RepoURL       string `json:"repo_url"`       // 仓库的URL
	TargetPath    string `json:"target_path"`    // 仓库克隆的目标路径
	WebhookPort   string `json:"webhook_port"`   // Webhook 监听的端口
	StaticPort    string `json:"static_port"`    // 静态文件服务监听的端口
	StaticPath    string `json:"static_path"`    // 静态内容服务的路径
}

// loadConfig 从文件中读取配置，如果文件不存在则创建默认配置
func loadConfig(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 默认配置
		defaultConfig := Config{
			RepoURL:     "https://github.com/yourusername/yourrepo.git",
			TargetPath:  "./repo",
			WebhookPort: "8081",                    // Webhook 服务端口
			StaticPort:  "8080",                    // 静态文件服务端口
			StaticPath:  "./repo",       // 仓库中的默认静态文件夹
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
		// 退出程序
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

// cloneRepo 克隆仓库到目标路径
func cloneRepo(config *Config) error {
	_, err := git.PlainClone(config.TargetPath, false, &git.CloneOptions{
		URL: config.RepoURL,
	})
	return err
}

// pullRepo 从远程仓库拉取最新的更改
func pullRepo(config *Config) error {
	repo, err := git.PlainOpen(config.TargetPath)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// 拉取最新的更改
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err.Error() == "already up-to-date" {
		log.Println("仓库已经是最新状态")
		return nil
	}
    if err == nil {
        log.Println("仓库已更新")
    }
	return err
}

// webhookHandler 处理Webhook请求以拉取最新的更改
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

// serveStaticFiles 设置HTTP文件服务器以提供静态文件
func serveStaticFiles(staticPath, port string) {
	// 从指定目录提供静态文件
	fs := http.FileServer(http.Dir(staticPath))
	http.Handle("/", fs)
	log.Printf("静态文件服务器监听在端口 %s，提供 %s 的内容", port, staticPath)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("启动静态文件服务器时出错: %v", err)
	}
}

func serveWebhook(config *Config) {
	http.HandleFunc("/webhook", webhookHandler(config))
	log.Printf("Webhook 服务器监听在端口 %s", config.WebhookPort)
	if err := http.ListenAndServe(":"+config.WebhookPort, nil); err != nil {
		log.Fatalf("启动 Webhook 服务器时出错: %v", err)
	}
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("加载配置时出错: %v", err)
	}
	log.Println("载入配置")
	log.Println("同步自：", config.RepoURL)

	// 检查仓库是否存在，如果不存在则克隆
	if _, err := os.Stat(config.TargetPath); os.IsNotExist(err) {
		log.Println("未找到仓库，正在克隆...")
		if err := cloneRepo(config); err != nil {
			log.Fatalf("克隆仓库时出错: %v", err)
		}
	}

	// 启动静态文件服务器和 Webhook 服务器
	go serveStaticFiles(config.StaticPath, config.StaticPort) // 静态文件服务器
	serveWebhook(config)                                     // Webhook 服务器
}
