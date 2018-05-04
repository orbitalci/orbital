package ssh

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/basher"
	"github.com/shankj3/ocelot/common/helpers/sshhelper"
	"github.com/shankj3/ocelot/common/testutil"
	"github.com/shankj3/ocelot/models"
)

func SetupSSHBuilder(t *testing.T, sshPort int, servicePort string) (bldr build.Builder, ctx context.Context, cancel func(), tarRm func(*testing.T), dockerClean func()) {
	if testing.Short() {
		t.Skip("skipping ssh setup test due to -short flag being set and this being a docker-heavy test")
	}
	rmTar := testutil.TarTemplates(t, "../builder/ssh/test-fixtures/werker_files.tar", "../../template/")
	//defer rmTar(t)
	go testutil.CreateDoThingsWebServer("./test-fixtures/werker_files.tar", servicePort)

	cleanup, ctx := sshhelper.CreateSSHDockerContainer(t, fmt.Sprintf("%d", sshPort))
	//defer cleanup()
	bshr, err := basher.NewBasher("", "", "docker.for.mac.localhost", "/tmp")
	if err != nil {
		t.Error(err)
		return
	}
	sshfacts := &models.SSHFacts{
		User: "root",
		Host: "localhost",
		Port: sshPort,
		KeyFP: "./test-fixtures/docker_id_rsa",
	}

	werkerFacts := &models.WerkerFacts{
		Uuid: uuid.New(),
		WerkerType: models.SSH,
		LoopbackIp: "docker.for.mac.localhost",
		RegisterIP: "docker.for.mac.localhost",
		ServicePort: servicePort,
		Dev: true,
		Ssh: sshfacts,
	}
	ssh, err := NewSSHBuilder(bshr, werkerFacts)
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel = context.WithCancel(ctx)
	return ssh, ctx, cancel, rmTar, cleanup
}
