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

const configPath = "etc/config.json"

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
	cfg, err := config.LoadConfig(configPath)
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
		log.Println("使用 AB 分区策略: 活动分区", cfg.ActivePartition)
	} else {
		log.Println("未启用Git LFS")
	}

	if cfg.WebhookSecret != "" {
		log.Println("已启用Webhook安全验证")
	} else {
		log.Println("警告: 未启用Webhook安全验证，建议在配置中设置webhook_secret")
	}

	// 获取活动分区路径
	activePath := cfg.GetActiveTargetPath()
	log.Printf("当前活动分区: %s", activePath)

	if _, err := os.Stat(activePath); os.IsNotExist(err) {
		log.Println("未找到仓库，正在克隆...")
		if err := repo.CloneRepoToPath(cfg, activePath); err != nil {
			log.Fatalf("克隆仓库时出错: %v", err)
		}
	} else {
		log.Println("发现现有仓库")
		if cfg.UpdateOnStart {
			log.Println("检查仓库更新...")
			if err := repo.PullRepo(cfg); err != nil {
				log.Printf("更新仓库时出错: %v，将尝试重新克隆", err)
				// 如果更新失败，尝试删除并重新克隆
				if err := os.RemoveAll(activePath); err != nil {
					log.Fatalf("删除现有仓库失败: %v", err)
				}
				if err := repo.CloneRepoToPath(cfg, activePath); err != nil {
					log.Fatalf("重新克隆仓库时出错: %v", err)
				}
			}
		}
	}

	log.Printf("Git2Web 成功启动! 启动用时: %v", time.Since(server.StartTime))
	log.Printf("静态文件服务: http://IP:%s (从 %s 提供服务)", cfg.StaticPort, activePath)

	go server.ServeStaticFiles(activePath, cfg.StaticPort)
	go server.ServeWebhook(cfg, configPath)
	select {}
}
