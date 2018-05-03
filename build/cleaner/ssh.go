package cleaner

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models"
)

type SSHCleaner struct {}

func (d *SSHCleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	prefix := build.GetOcyPrefixFromWerkerType(models.SSH)
	cloneDir := build.GetCloneDir(prefix, id)
	logout <- []byte("Removing build directory " + cloneDir)
	if err := os.RemoveAll(cloneDir); err != nil {
		logout <- []byte("Could not remove build directory! Error: " + err.Error())
		return err
	}
	logout <- []byte("Successfully removed build directory.")
	return nil
}
