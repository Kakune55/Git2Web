package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	logo := `
	
 ██████╗ ██╗████████╗██████╗ ██╗    ██╗███████╗██████╗ 
██╔════╝ ██║╚══██╔══╝╚════██╗██║    ██║██╔════╝██╔══██╗
██║  ███╗██║   ██║    █████╔╝██║ █╗ ██║█████╗  ██████╔╝
██║   ██║██║   ██║   ██╔═══╝ ██║███╗██║██╔══╝  ██╔══██╗
╚██████╔╝██║   ██║   ███████╗╚███╔███╔╝███████╗██████╔╝
 ╚═════╝ ╚═╝   ╚═╝   ╚══════╝ ╚══╝╚══╝ ╚══════╝╚═════╝ 
	Version: 1.1.0

	`
	fmt.Println(logo)
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

	if _, err := os.Stat(config.TargetPath); os.IsNotExist(err) {
		log.Println("未找到仓库，正在克隆...")
		if err := cloneRepo(config); err != nil {
			log.Fatalf("克隆仓库时出错: %v", err)
		}
	}

	log.Print("正在启动 Web 服务器...")
	go serveStaticFiles(config.StaticPath, config.StaticPort)
	serveWebhook(config)
}
