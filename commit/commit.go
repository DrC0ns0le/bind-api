package commit

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	conf "github.com/DrC0ns0le/bind-api/config"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

const (
	directory = "output"
)

var (
	token       = flag.String("git.token", "", "git token, env: GIT_TOKEN")
	url         = flag.String("git.url", "", "git url, env: GIT_URL")
	commitName  = flag.String("git.name", "", "commit name, env: GIT_NAME")
	commitEmail = flag.String("git.email", "", "commit email, env: GIT_EMAIL")
)

var (
	authMethod transport.AuthMethod
)

func Init() error {
	// check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// set up authentication
	*token = conf.GetEnv("GIT_TOKEN", *token)
	if *token != "" {
		authMethod = &http.BasicAuth{
			Username: "token",
			Password: *token,
		}
	} else {
		var err error
		authMethod, err = ssh.NewSSHAgentAuth("git")
		if err != nil {
			return fmt.Errorf("failed to setup SSH auth: %w", err)
		}
	}

	// check if git is already cloned
	if _, err := os.Stat(directory + "/.git"); os.IsNotExist(err) {
		_, err = git.PlainClone(directory, false, &git.CloneOptions{
			Auth: authMethod,
			URL:  conf.GetEnv("GIT_URL", *url),
		})
		if err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	if err := Reset(); err != nil {
		return fmt.Errorf("failed to reset repository: %w", err)
	}

	log.Println("Git init successful.")
	return nil
}

// Commit all files and push to remote
func Push() error {

	// open repo
	r, err := git.PlainOpen(directory)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// get worktree
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// git add .
	_, err = w.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add files to repository: %w", err)
	}

	// git commit -m \"message\"
	commitMsg := fmt.Sprintf("api commit at %s", time.Now().Format(time.RFC3339))
	commit, err := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  conf.GetEnv("GIT_NAME", *commitName),
			Email: conf.GetEnv("GIT_EMAIL", *commitEmail),
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	_, err = r.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("failed to commit repository: %w", err)
	}

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       authMethod,
	})
	if err != nil {
		return fmt.Errorf("failed to push repository: %w", err)
	}

	return nil
}

// Undo all changes
func Reset() error {
	r, err := git.PlainOpen(directory)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the default remote
	remote, err := r.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	// Fetch latest changes first
	err = remote.Fetch(&git.FetchOptions{
		Auth:     authMethod,
		Force:    true,
		RefSpecs: []config.RefSpec{"refs/*:refs/*"},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	// Hard reset to remove all local changes
	err = w.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset repository: %w", err)
	}

	// Get current branch reference
	ref, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Reset to remote branch
	err = w.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: plumbing.NewHash("origin/" + ref.Name().Short()),
	})
	if err != nil {
		return fmt.Errorf("failed to reset to remote: %w", err)
	}

	return nil
}

// Check if staging
func Staging() (bool, error) {
	r, err := git.PlainOpen(directory)
	if err != nil {
		return false, fmt.Errorf("failed to open repository: %w", err)
	}
	w, err := r.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	if len(status) > 0 {
		return true, nil
	}
	return false, nil
}
