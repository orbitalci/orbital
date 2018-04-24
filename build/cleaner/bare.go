package cleaner

import (
	"context"

	"bitbucket.org/level11consulting/go-til/log"
)

type BareCleaner struct {}

func (d *BareCleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	log.Log().Info("machine build with id ", id, "is finished. there is no cleanup for it.")
	logout <- []byte("Finished Cleaning.")
	return nil
}
