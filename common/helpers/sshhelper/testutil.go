package sshhelper

import (
	"context"
	"os"
	"testing"

	"github.com/level11consulting/ocelot/common/testutil"
)

func CreateSSHDockerContainer(t *testing.T, forwardPort string) (cleanup func(), ctx context.Context) {
	var err error
	if testing.Short() {
		t.Skip("skipping docker container test because -short flag set")
	}
	sshMount := os.ExpandEnv("$PWD/test-fixtures/docker_id_rsa.pub") + ":/root/.ssh/authorized_keys"
	configMount := os.ExpandEnv("$PWD/test-fixtures/sshd_config") + ":/etc/ssh/sshd_config:ro"
	ctx, _ = context.WithCancel(context.Background())
	cleanup, err = testutil.DockerCreateExec(t, ctx, "panubo/sshd", []string{forwardPort + ":22"}, sshMount, configMount)
	if err != nil {
		t.Fatal("couldn't create container", err.Error())
		return
	}
	return cleanup, ctx
}
