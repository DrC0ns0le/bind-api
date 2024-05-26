package commit

import (
	"fmt"
	"log"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

const (
	url       = "git@github.com:DrC0ns0le/internal-bind-config.git"
	directory = "output"
)

func init() {

	// check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, 0755)
	}

	// check if git is already cloned
	if _, err := os.Stat(directory + "/.git"); os.IsNotExist(err) {
		authMethod, err := ssh.NewSSHAgentAuth("git")

		if err != nil {
			panic(err)
		}

		_, err = git.PlainClone(directory, false, &git.CloneOptions{
			Auth: authMethod,
			URL:  url,
		})

		if err != nil {
			panic(err)
		}
	}

	Reset()

	log.Println("Git init successful.")
}

// Commit all files and push to remote
func Push() error {

	// open repo
	r, err := git.PlainOpen(directory)
	if err != nil {
		panic(err)
	}

	// get worktree
	w, err := r.Worktree()
	if err != nil {
		panic(err)
	}

	// git add .
	_, err = w.Add(".")
	if err != nil {
		return err
	}

	// git commit -m \"message\"
	commitMsg := fmt.Sprintf("api commit at %s", time.Now().Format(time.RFC3339))
	commit, err := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Bind Bot",
			Email: "bind.bot@leejacksonz.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	_, err = r.CommitObject(commit)
	if err != nil {
		return err
	}

	err = r.Push(&git.PushOptions{})
	if err != nil {
		return err
	}

	return nil
}

// Undo all changes
func Reset() error {

	r, err := git.PlainOpen(directory)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return err
		}
	}

	// reset to the latest commit
	err = w.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash("master"),
		Force: true,
	})
	if err != nil {
		return err
	}

	return nil
}

// Check if staging
func Staging() (bool, error) {
	r, err := git.PlainOpen(directory)
	if err != nil {
		return false, err
	}
	w, err := r.Worktree()
	if err != nil {
		return false, err
	}

	status, err := w.Status()
	if err != nil {
		return false, err
	}

	if len(status) > 0 {
		return true, nil
	}
	return false, nil
}
