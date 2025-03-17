package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"git2Web/config"
	"git2Web/repo"
	"git2Web/security"
)

var StartTime time.Time

func init() {
	StartTime = time.Now()
}

func webhookHandler(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("收到Webhook请求")
		
		// 验证请求
		if !security.ValidateWebhook(r, config.WebhookSecret) {
			http.Error(w, "未授权的请求", http.StatusUnauthorized)
			log.Println("Webhook 验证失败")
			return
		}
		
		err := repo.PullRepo(config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("拉取仓库时出错: %v", err)
			return
		}
		fmt.Fprintln(w, "仓库成功更新")
	}
}

// 健康检查端点
func healthCheckHandler(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"status":  "healthy",
			"version": config.Version,
			"uptime":  time.Since(StartTime).String(),
			"repo": map[string]string{
				"url":        config.RepoURL,
				"targetPath": config.TargetPath,
			},
		}
		
		// 检查目标目录是否存在
		_, err := os.Stat(config.TargetPath)
		info["repoExists"] = err == nil
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}

func ServeStaticFiles(staticPath, port string) {
	fs := http.FileServer(http.Dir(staticPath))
	http.Handle("/", fs)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("启动静态文件服务器时出错: %v", err)
	}
}

func ServeWebhook(config *config.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookHandler(config))
	mux.HandleFunc("/health", healthCheckHandler(config))
	
	log.Printf("Webhook 服务: http://IP:%s/webhook", config.WebhookPort)
	log.Printf("健康检查端点: http://IP:%s/health", config.WebhookPort)
	
	server := &http.Server{
		Addr:         ":" + config.WebhookPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("启动 Webhook 服务器时出错: %v", err)
	}
}
