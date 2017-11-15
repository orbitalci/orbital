package getgit

import (
	"io/ioutil"
	"os"
	"os/exec"
)

//todo: use bitbucket token to clone with git

// Shallow Clones Repostory at a depth of 100 commits into a temporary directory
// If there are any errors in the clone, the temp directory is erased.
// Returns temporary directory filepath and any errors received
func ShallowCloneRepo(repourl string) (string, error) {
	dir, err := ioutil.TempDir("", "gitrepo")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "clone", "--depth=100", repourl, dir)
	err = cmd.Run()
	if err != nil {
		os.Remove(dir)
		return "", err
	}
	return dir, nil
}

// Checks out specific commit of repository, returns any error from command.
func CheckOutRepoHash(repo_direc string, git_hash string) error {
	cmd := exec.Command("git", "checkout", git_hash)
	cmd.Dir = repo_direc
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
