package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// 全局变量记录程序启动时间
var startTime time.Time

func main() {
	startTime = time.Now()
	
	logo := `
	
 ██████╗ ██╗████████╗██████╗ ██╗    ██╗███████╗██████╗ 
██╔════╝ ██║╚══██╔══╝╚════██╗██║    ██║██╔════╝██╔══██╗
██║  ███╗██║   ██║    █████╔╝██║ █╗ ██║█████╗  ██████╔╝
██║   ██║██║   ██║   ██╔═══╝ ██║███╗██║██╔══╝  ██╔══██╗
╚██████╔╝██║   ██║   ███████╗╚███╔███╔╝███████╗██████╔╝
 ╚═════╝ ╚═╝   ╚═╝   ╚══════╝ ╚══╝╚══╝ ╚══════╝╚═════╝ 
	Version: ` + AppVersion + `

	`
	fmt.Println(logo)
	
	log.Printf("Git2Web 启动中，版本: %s, 系统: %s/%s", 
		AppVersion, runtime.GOOS, runtime.GOARCH)
		
	log.Println("载入配置...")
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("加载配置时出错: %v", err)
	}

	// 初始化日志系统
	if err := initLogging(config); err != nil {
		log.Fatalf("初始化日志系统时出错: %v", err)
	}
	log.Println("日志目录：", config.LogFilePath)


	log.Println("同步自：", config.RepoURL)
	if config.RepoAuth.Enabled {
		log.Println("已启用身份验证")
	} else {
		log.Println("未启用身份验证")
	}

	if config.LfsEnabled {
		log.Println("已启用Git LFS")
	} else {
		log.Println("未启用Git LFS")
	}
	
	if config.WebhookSecret != "" {
		log.Println("已启用Webhook安全验证")
	} else {
		log.Println("警告: 未启用Webhook安全验证，建议在配置中设置webhook_secret")
	}

	if _, err := os.Stat(config.TargetPath); os.IsNotExist(err) {
		log.Println("未找到仓库，正在克隆...")
		if err := cloneRepo(config); err != nil {
			log.Fatalf("克隆仓库时出错: %v", err)
		}
	} else {
		log.Println("发现现有仓库，检查更新...")
		if err := pullRepo(config); err != nil {
			log.Printf("更新仓库时出错: %v，将尝试重新克隆", err)
			// 如果更新失败，尝试删除并重新克隆
			if err := os.RemoveAll(config.TargetPath); err != nil {
				log.Fatalf("删除现有仓库失败: %v", err)
			}
			if err := cloneRepo(config); err != nil {
				log.Fatalf("重新克隆仓库时出错: %v", err)
			}
		}
	}

	log.Printf("Git2Web 成功启动! 运行时间: %v", time.Since(startTime))
	log.Printf("静态文件服务: http://IP:%s", config.StaticPort)
	
	go serveStaticFiles(config.StaticPath, config.StaticPort)
	go serveWebhook(config)
	select {}
}
