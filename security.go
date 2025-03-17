package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
)

// validateWebhook 验证来自 GitHub/GitLab 的 Webhook 请求
func validateWebhook(r *http.Request, secret string) bool {
	if secret == "" {
		// 如果未配置 secret，则不进行验证
		return true
	}

	// GitHub 使用 X-Hub-Signature-256 头，GitLab 使用 X-Gitlab-Token 头
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		signature = r.Header.Get("X-Gitlab-Token")
	}

	if signature == "" {
		log.Println("警告: 没有找到签名头")
		return false
	}

	// 对于 GitHub，验证 HMAC
	if r.Header.Get("X-GitHub-Event") != "" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("读取请求体失败: %v", err)
			return false
		}
		// 重置 body，以便后续处理可以再次读取
		r.Body = io.NopCloser(io.MultiReader(io.NopCloser(io.MultiReader()), io.NopCloser(io.MultiReader())))

		// 计算 HMAC
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedMAC := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		return hmac.Equal([]byte(signature), []byte(expectedMAC))
	}

	// 对于 GitLab，直接比较 token
	return signature == secret
}
