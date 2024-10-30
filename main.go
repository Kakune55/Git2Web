package main

import (
	"log"
	"os"
)

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
