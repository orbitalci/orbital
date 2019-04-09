package ssh

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/build/basher"
	"github.com/level11consulting/ocelot/common/helpers/sshhelper"
	"github.com/level11consulting/ocelot/common/testutil"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
)

func createSSHBuilderNoInit(t *testing.T, sshPort int, prefixdir string) (context.Context, *SSH, func()) {
	cleanup, ctx := sshhelper.CreateSSHDockerContainer(t, fmt.Sprintf("%d", sshPort))
	bshr, err := basher.NewBasher("", "", "docker.for.mac.localhost", prefixdir)
	if err != nil {
		t.Fatal(err)
	}
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	path := filepath.Join(dir, "test-fixtures")
	sshfacts := &models.SSHFacts{
		User:  "root",
		Host:  "localhost",
		Port:  sshPort,
		KeyFP: path + "/docker_id_rsa",
	}
	werkerFacts := &models.WerkerFacts{
		Uuid:        uuid.New(),
		WerkerType:  models.SSH,
		LoopbackIp:  "docker.for.mac.localhost",
		RegisterIP:  "docker.for.mac.localhost",
		ServicePort: "",
		Dev:         true,
		Ssh:         sshfacts,
	}
	ssh, err := NewSSHBuilder(bshr, werkerFacts)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, ssh.(*SSH), cleanup
}

func SetupSSHBuilderNoTempl(t *testing.T, sshPort int, testHash string, prefixdir string) (ctx context.Context, sh *SSH, dockerClean func()) {
	ctx, ssh, cleanup := createSSHBuilderNoInit(t, sshPort, prefixdir)
	logt := make(chan []byte)
	done := make(chan bool)
	var out string
	go func() {
		for i := range logt {
			out += string(i) + "\n"
		}
		close(done)
	}()
	res := ssh.Init(ctx, testHash, logt)
	close(logt)
	<-done
	//defer ssh.Close()
	if res.Status != pb.StageResultVal_PASS {
		t.Log(out)
		cleanup()
		t.Fatal("should pass, error is: ", res.Error)
	}
	expected := "Successfully established ssh connection " + models.CHECKMARK
	if res.Messages[0] != expected {
		t.Log(out)
		cleanup()
		t.Fatal(test.StrFormatErrors("init messages", expected, res.Messages[0]))
	}
	return ctx, ssh, cleanup
}

func SetupSSHBuilder(t *testing.T, sshPort int, servicePort string) (bldr build.Builder, ctx context.Context, cancel func(), tarRm func(*testing.T), dockerClean func()) {
	if testing.Short() {
		t.Skip("skipping ssh setup test due to -short flag being set and this being a docker-heavy test")
	}
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "template")
	t.Log("TEMPLATE DIR IS...." + dir)
	rmTar := testutil.TarTemplates(t, "../builder/ssh/test-fixtures/werker_files.tar", dir)
	//defer rmTar(t)
	go testutil.CreateDoThingsWebServer("./test-fixtures/werker_files.tar", servicePort)

	cleanup, ctx := sshhelper.CreateSSHDockerContainer(t, fmt.Sprintf("%d", sshPort))
	//defer cleanup()
	bshr, err := basher.NewBasher("", "", "docker.for.mac.localhost", "/tmp")
	if err != nil {
		t.Error(err)
		return
	}
	dir = filepath.Dir(filename)
	path := filepath.Join(dir, "test-fixtures")
	sshfacts := &models.SSHFacts{
		User:  "root",
		Host:  "localhost",
		Port:  sshPort,
		KeyFP: path + "/docker_id_rsa",
	}

	werkerFacts := &models.WerkerFacts{
		Uuid:        uuid.New(),
		WerkerType:  models.SSH,
		LoopbackIp:  "docker.for.mac.localhost",
		RegisterIP:  "docker.for.mac.localhost",
		ServicePort: servicePort,
		Dev:         true,
		Ssh:         sshfacts,
	}
	ssh, err := NewSSHBuilder(bshr, werkerFacts)
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel = context.WithCancel(ctx)
	time.Sleep(5)
	return ssh, ctx, cancel, rmTar, cleanup
}
