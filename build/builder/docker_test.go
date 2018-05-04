package builder

import (
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models/pb"
	"golang.org/x/net/context"

	"testing"
)

// test that in docker, can run the InstallPackageDeps to multiple image types
func TestDockerBasher_InstallPackageDeps(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping due to -short flag being set")
	}
	ctx := context.Background()
	alpine, cleanupFunc := CreateLivingDockerContainer(t, ctx, "alpine:latest")
	defer cleanupFunc(t)
	su := build.InitStageUtil("alpineTest")
	logout := make(chan []byte, 10000)
	result := alpine.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, alpine.InstallPackageDeps(), logout)
	if result.Status == pb.StageResultVal_FAIL {
		t.Error("couldn't download deps! oh nuuu!")
	}
	t.Log(result.Status)
	t.Log(string(<-logout))
	testDeps := []string{"/bin/sh", "-c", "command -v openssl && command -v bash && command -v zip && command -v wget && command -v python"}
	result = alpine.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, testDeps, logout)
	if result.Status == pb.StageResultVal_FAIL {
		t.Error("deps not found! oh nuuu!")
	}
	t.Log(result.Status)
	t.Log(string(<-logout))

}
