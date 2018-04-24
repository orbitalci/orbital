package sshhelper

import (
	"errors"
	"os"

	"github.com/kevinburke/ssh_config"
)

func GetPrivateKeyFileFromConfig(sshConfigPath string) (string, error) {
	var key string
	var err error
	var f *os.File
	f, err = os.Open(sshConfigPath)
	if err != nil {
		return "", err
	}
	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return "", err
	}
	key, err = cfg.Get("default", "IdentityFile")
	if key == "" && err == nil {
		err = errors.New("IdentityFile not found in ssh config at path: " + sshConfigPath)
	}
	return key, err
}