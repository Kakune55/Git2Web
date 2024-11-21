package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// cloneRepo 克隆仓库
func cloneRepo(config *Config) error {
	cloneOptions := &git.CloneOptions{
		URL: config.RepoURL,
	}

	// 如果需要认证，设置认证信息
	if config.RepoAuth.Enabled {
		cloneOptions.Auth = &http.BasicAuth{
			Username: config.RepoAuth.Email,
			Password: config.RepoAuth.Password,
		}
	}

	// 克隆仓库
	log.Printf("开始克隆仓库: %s 到路径: %s", config.RepoURL, config.TargetPath)
	_, err := git.PlainClone(config.TargetPath, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("克隆仓库失败: %w", err)
	}

	log.Println("仓库克隆完成")

	// 如果启用了 LFS，执行 LFS 拉取
	if config.LfsEnabled {
		if err := updateGitLFS(config.TargetPath, config); err != nil {
			return fmt.Errorf("git LFS 拉取失败: %w", err)
		}
	}

	return nil
}

// pullRepo 拉取更新
func pullRepo(config *Config) error {
    // 如果启用了 LFS，则直接删除现有仓库并重新克隆
    if config.LfsEnabled {
        log.Println("启用 LFS，直接删除现有仓库并重新克隆")
        if err := os.RemoveAll(config.TargetPath); err != nil {
            return fmt.Errorf("删除现有仓库失败: %w", err)
        }
        return cloneRepo(config)
    }

    // 打开已有的仓库
    repo, err := git.PlainOpen(config.TargetPath)
    if err != nil {
        return fmt.Errorf("打开仓库失败: %w", err)
    }

    // 获取工作树
    w, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("获取工作树失败: %w", err)
    }

    // 设置拉取选项
    pullOptions := &git.PullOptions{
        RemoteName: "origin",
    }

    if config.RepoAuth.Enabled {
        pullOptions.Auth = &http.BasicAuth{
            Username: config.RepoAuth.Email,
            Password: config.RepoAuth.Password,
        }
    }

    // 执行拉取操作
    log.Println("开始拉取仓库更新")
    err = w.Pull(pullOptions)
    if err != nil {
        if err.Error() == "already up-to-date" {
            log.Println("仓库已经是最新状态")
            return nil
        }
        return fmt.Errorf("拉取仓库更新失败: %w", err)
    }

    log.Println("仓库更新完成")

    return nil
}

// updateGitLFS 使用命令行工具拉取 Git LFS 文件
func updateGitLFS(targetPath string, config *Config) error {
	log.Println("开始更新 Git LFS 文件")
	cmd := exec.Command("git", "lfs", "pull")
	cmd.Dir = targetPath

	// 设置环境变量以传递认证信息
	if config.RepoAuth.Enabled {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_USERNAME=%s", config.RepoAuth.Email))
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_PASSWORD=%s", config.RepoAuth.Password))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git LFS 更新失败: %s, 错误信息: %s", err, string(output))
	}
	log.Println("Git LFS 更新完成")
	return nil
}
