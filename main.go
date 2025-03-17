package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"git2Web/config"
	"git2Web/logger"
	"git2Web/repo"
	"git2Web/server"
)

func main() {
	server.StartTime = time.Now()
	
	logo := `
	
 ██████╗ ██╗████████╗██████╗ ██╗    ██╗███████╗██████╗ 
██╔════╝ ██║╚══██╔══╝╚════██╗██║    ██║██╔════╝██╔══██╗
██║  ███╗██║   ██║    █████╔╝██║ █╗ ██║█████╗  ██████╔╝
██║   ██║██║   ██║   ██╔═══╝ ██║███╗██║██╔══╝  ██╔══██╗
╚██████╔╝██║   ██║   ███████╗╚███╔███╔╝███████╗██████╔╝
 ╚═════╝ ╚═╝   ╚═╝   ╚══════╝ ╚══╝╚══╝ ╚══════╝╚═════╝ 
	Version: ` + config.AppVersion + `

	`
	fmt.Println(logo)
	
	log.Printf("Git2Web 启动中，版本: %s, 系统: %s/%s", 
		config.AppVersion, runtime.GOOS, runtime.GOARCH)
		
	log.Println("载入配置...")
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("加载配置时出错: %v", err)
	}

	// 初始化日志系统
	if err := logger.InitLogging(cfg); err != nil {
		log.Fatalf("初始化日志系统时出错: %v", err)
	}
	log.Println("日志目录：", cfg.LogFilePath)


	log.Println("同步自：", cfg.RepoURL)
	if cfg.RepoAuth.Enabled {
		log.Println("已启用身份验证")
	} else {
		log.Println("未启用身份验证")
	}

	if cfg.LfsEnabled {
		log.Println("已启用Git LFS")
	} else {
		log.Println("未启用Git LFS")
	}
	
	if cfg.WebhookSecret != "" {
		log.Println("已启用Webhook安全验证")
	} else {
		log.Println("警告: 未启用Webhook安全验证，建议在配置中设置webhook_secret")
	}

	if _, err := os.Stat(cfg.TargetPath); os.IsNotExist(err) {
		log.Println("未找到仓库，正在克隆...")
		if err := repo.CloneRepo(cfg); err != nil {
			log.Fatalf("克隆仓库时出错: %v", err)
		}
	} else {
		log.Println("发现现有仓库，检查更新...")
		if err := repo.PullRepo(cfg); err != nil {
			log.Printf("更新仓库时出错: %v，将尝试重新克隆", err)
			// 如果更新失败，尝试删除并重新克隆
			if err := os.RemoveAll(cfg.TargetPath); err != nil {
				log.Fatalf("删除现有仓库失败: %v", err)
			}
			if err := repo.CloneRepo(cfg); err != nil {
				log.Fatalf("重新克隆仓库时出错: %v", err)
			}
		}
	}

	log.Printf("Git2Web 成功启动! 启动用时: %v", time.Since(server.StartTime))
	log.Printf("静态文件服务: http://IP:%s", cfg.StaticPort)
	
	go server.ServeStaticFiles(cfg.StaticPath, cfg.StaticPort)
	go server.ServeWebhook(cfg)
	select {}
}
