package main

import (
	"log"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func cloneRepo(config *Config) error {
	if config.RepoAuth["enabled"] == "1" {
		_, err := git.PlainClone(config.TargetPath, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: config.RepoAuth["username"],
				Password: config.RepoAuth["password"],
			},
			URL: config.RepoURL,
		})
		return err
	} else {
		_, err := git.PlainClone(config.TargetPath, false, &git.CloneOptions{
			URL: config.RepoURL,
		})
		return err
	}
}

func pullRepo(config *Config) error {
	repo, err := git.PlainOpen(config.TargetPath)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err.Error() == "already up-to-date" {
		log.Println("仓库已经是最新状态")
		return nil
	}
	return err
}
