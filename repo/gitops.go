package repo

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"

	"git2Web/config"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// GetBranchInfo 获取当前分支信息
func GetBranchInfo(repoPath string) {
	// 打开当前的 Git 仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatalf("无法打开仓库: %v", err)
	}

	// 获取当前 HEAD 的引用
	headRef, err := repo.Head()
	if err != nil {
		log.Fatalf("无法获取 HEAD: %v", err)
	}

	// 输出当前节点的信息
	log.Printf("当前 HEAD Commit: %s\n", headRef.Hash().String())

	// 判断 HEAD 是否指向分支
	if headRef.Name().IsBranch() {
		log.Printf("当前分支: %s\n", headRef.Name().Short())
	} else {
		log.Println("当前是一个分离的 HEAD 状态 (Detached HEAD)")
	}

	commit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		log.Fatalf("无法获取 Commit 对象: %v", err)
	}
	log.Printf("当前的提交信息: %s", commit.Message)
}

// CloneRepo 克隆仓库到默认路径
func CloneRepo(config *config.Config) error {
	return CloneRepoToPath(config, config.GetActiveTargetPath())
}

// CloneRepoToPath 克隆仓库到指定路径
func CloneRepoToPath(config *config.Config, targetPath string) error {
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

	// 确保目标路径存在
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 克隆仓库
	log.Printf("开始克隆仓库: %s 到路径: %s", config.RepoURL, targetPath)
	_, err := git.PlainClone(targetPath, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("克隆仓库失败: %w", err)
	}

	log.Println("仓库克隆完成")

	// 如果启用了 LFS，执行 LFS 拉取
	if config.LfsEnabled {
		if err := updateGitLFS(targetPath, config); err != nil {
			return fmt.Errorf("git LFS 拉取失败: %w", err)
		}
	}
	
	GetBranchInfo(targetPath)

	return nil
}

// PullRepo 拉取更新
func PullRepo(config *config.Config) error {
    targetPath := config.GetActiveTargetPath()
    
    // 如果启用了 LFS，可以在这里处理 AB 分区策略
    // 但实际实现在 server 包中的 webhookHandler 中
    if config.LfsEnabled {
        log.Println("已启用 LFS 和 AB 分区策略，通过 webhook 处理...")
        return nil
    }

    // 打开已有的仓库
    repo, err := git.PlainOpen(targetPath)
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
	GetBranchInfo(targetPath)

    return nil
}

// updateGitLFS 使用命令行工具拉取 Git LFS 文件
func updateGitLFS(targetPath string, config *config.Config) error {
	log.Println("开始更新 Git LFS 文件")

	// 如果需要认证，临时设置带认证信息的远程 URL
	originURL := config.RepoURL
	authURL := originURL
	needRestore := false

	if config.RepoAuth.Enabled {
		needRestore = true
		authURL = originURL
		if len(config.RepoAuth.Email) > 0 && len(config.RepoAuth.Password) > 0 {
			const prefix = "https://"
			if len(originURL) > len(prefix) && originURL[:len(prefix)] == prefix {
				// 对用户名和密码做 URL 编码
				encodedUser := url.QueryEscape(config.RepoAuth.Email)
				encodedPass := url.QueryEscape(config.RepoAuth.Password)
				authURL = fmt.Sprintf("https://%s:%s@%s", encodedUser, encodedPass, originURL[len(prefix):])
			}
		}
		// 设置远程 URL
		cmdSet := exec.Command("git", "remote", "set-url", "origin", authURL)
		cmdSet.Dir = targetPath
		if output, err := cmdSet.CombinedOutput(); err != nil {
			return fmt.Errorf("设置带认证信息的远程 URL 失败: %s, 错误信息: %s", err, string(output))
		}
	}

	// 执行 git lfs pull
	cmd := exec.Command("git", "lfs", "pull")
	cmd.Dir = targetPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 恢复远程 URL
		if needRestore {
			cmdRestore := exec.Command("git", "remote", "set-url", "origin", originURL)
			cmdRestore.Dir = targetPath
			cmdRestore.CombinedOutput()
		}
		return fmt.Errorf("git LFS 更新失败: %s, 错误信息: %s", err, string(output))
	}

	// 恢复远程 URL
	if needRestore {
		cmdRestore := exec.Command("git", "remote", "set-url", "origin", originURL)
		cmdRestore.Dir = targetPath
		cmdRestore.CombinedOutput()
	}

	log.Println("Git LFS 更新完成")
	return nil
}
