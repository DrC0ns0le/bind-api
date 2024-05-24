package git

import (
	"fmt"
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

const (
	url       = "https://github.com/DrC0ns0le/internal-bind-config.git"
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

		fmt.Println("Hi!")
		_, err = git.PlainClone(directory, false, &git.CloneOptions{
			Auth:     authMethod,
			URL:      url,
			Progress: os.Stdout,
		})

		if err != nil {
			panic(err)
		}
	}

	// git pull
	r, err := git.PlainOpen(directory)
	if err != nil {
		panic(err)
	}

	w, err := r.Worktree()
	if err != nil {
		panic(err)
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			panic(err)
		}
	}

	// reset to the latest commit
	err = w.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash("master"),
		Force: true,
	})
	if err != nil {
		panic(err)
	}

	log.Println("Git init successful.")
}
