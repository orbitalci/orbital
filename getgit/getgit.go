package getgit 

import (
    "os"
    "log"
    "os/exec"
    "io/ioutil"
)


func ShallowCloneRepo(repourl string) (string, error) {
    dir, err := ioutil.Tempdir("", "gitrepo")
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

func CheckOutRepoHash(repo_direc string, git_hash string) error {
    cmd = exec.Command("git", "checkout", git_hash)
    cmd.Dir = repo_direc
    err := cmd.Run()
    if err != nil {
        return err
    }
    return nil
}