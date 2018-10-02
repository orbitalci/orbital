package cleaner

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models"
)

type ExecCleaner struct {
	prefix string
}

func NewExecCleaner() *ExecCleaner {
	return &ExecCleaner{prefix: build.GetOcyPrefixFromWerkerType(models.Exec)}
}

func (e *ExecCleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	var err error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if id == "" {
		return errors.New("id cannot be empty")
	}
	cloneDir := build.GetCloneDir(e.prefix, id)
	if logout != nil {
		logout <- []byte("removing build directory " + cloneDir)
	}
	err = os.RemoveAll(cloneDir)

	if logout != nil {
		if err != nil {
			failedCleaning.WithLabelValues("exec").Inc()
			logout <- []byte("error removing build directory: " + err.Error())
		} else {
			logout <- []byte("successfully removed build directory.")
		}
	}
	// if the context has been cancelled, then it was killed, as this deferred cleanup function is higher in the stack than the deferred cancel in (*launcher).makeitso
	if ctx.Err() == context.Canceled && logout != nil {
		logout <- []byte("//////////REDRUM////////REDRUM////////REDRUM/////////")
	}
	return nil
}
