package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"git2Web/config"
	"git2Web/repo"
	"git2Web/security"
)

var StartTime time.Time
var staticServer *http.Server
var staticServerMutex sync.Mutex

func init() {
	StartTime = time.Now()
}

func webhookHandler(config *config.Config, configPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("收到Webhook请求")
		updateStartTime := time.Now()
		
		// 验证请求
		if !security.ValidateWebhook(r, config.WebhookSecret) {
			http.Error(w, "未授权的请求", http.StatusUnauthorized)
			log.Println("Webhook 验证失败")
			return
		}
		
		// 使用 AB 分区策略更新仓库
		if config.LfsEnabled {
			log.Println("使用 AB 分区策略更新 LFS 仓库")
			inactivePath := config.GetInactiveTargetPath()
			
			// 如果非激活分区存在，则先删除
			if _, err := os.Stat(inactivePath); err == nil {
				log.Printf("清理非激活分区: %s", inactivePath)
				if err := os.RemoveAll(inactivePath); err != nil {
					http.Error(w, fmt.Sprintf("清理非激活分区失败: %v", err), http.StatusInternalServerError)
					log.Printf("清理非激活分区失败: %v", err)
					return
				}
			}
			
			// 克隆到非激活分区
			log.Printf("开始克隆到非激活分区: %s", inactivePath)
			if err := repo.CloneRepoToPath(config, inactivePath); err != nil {
				http.Error(w, fmt.Sprintf("克隆仓库失败: %v", err), http.StatusInternalServerError)
				log.Printf("克隆仓库失败: %v", err)
				return
			}
			
			// 切换活动分区
			log.Println("切换活动分区")
			config.SwitchActivePartition()
			
			// 保存配置更改
			if err := config.SaveConfig(configPath); err != nil {
				log.Printf("保存配置文件失败: %v", err)
				// 但不阻止服务切换
			}
			
			// 重启静态文件服务
			log.Println("重启静态文件服务到新分区")
			RestartStaticServer(config.GetActiveTargetPath(), config.StaticPort)
			
			fmt.Fprintln(w, "仓库成功更新并切换服务到新版本,用时:", time.Since(updateStartTime).String())
			log.Println("仓库成功更新并切换服务到新版本,用时:", time.Since(updateStartTime).String())
		} else {
			// 非 LFS 仓库使用常规更新方式
			err := repo.PullRepo(config)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Printf("拉取仓库时出错: %v", err)
				return
			}
			fmt.Fprintln(w, "仓库成功更新,用时:", time.Since(updateStartTime).String())
			log.Println("仓库成功更新,用时:", time.Since(updateStartTime).String())
		}
	}
}

// 健康检查端点
func healthCheckHandler(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activePath := config.GetActiveTargetPath()
		info := map[string]interface{}{
			"status":  "healthy",
			"version": config.Version,
			"uptime":  time.Since(StartTime).String(),
			"repo": map[string]string{
				"url":          config.RepoURL,
				"active_path":  activePath,
				"partition":    config.ActivePartition,
			},
		}
		
		// 检查目标目录是否存在
		_, err := os.Stat(activePath)
		info["repoExists"] = err == nil
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}

// RestartStaticServer 重启静态文件服务器，指向新目录
func RestartStaticServer(staticPath, port string) {
	staticServerMutex.Lock()
	defer staticServerMutex.Unlock()
	
	// 如果服务已存在，先关闭
	if staticServer != nil {
		log.Println("正在关闭现有静态文件服务器...")

		// 创建一个带超时的上下文
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		// 创建一个带超时的上下文用于关闭服务
		go func() {
			if err := staticServer.Shutdown(ctx); err != nil {
				log.Printf("关闭静态文件服务器时出错: %v", err)
			}
		}()
		// 等待一小段时间确保服务关闭
		time.Sleep(500 * time.Millisecond)
	}
	
	// 启动新服务
	log.Printf("启动静态文件服务器，路径: %s, 端口: %s", staticPath, port)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(staticPath)))
	
	staticServer = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	go func() {
		if err := staticServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("静态文件服务器错误: %v", err)
		}
	}()
}

func ServeStaticFiles(staticPath, port string) {
	RestartStaticServer(staticPath, port)
}

func ServeWebhook(config *config.Config, configPath string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookHandler(config, configPath))
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
