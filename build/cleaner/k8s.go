package cleaner

import (
	"context"
)

type K8Cleaner struct{}

func (k *K8Cleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	//TODO: implement this when the time comes
	return nil
}
