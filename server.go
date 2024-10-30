package main

import (
	"fmt"
	"log"
	"net/http"
)

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
