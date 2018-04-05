package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/integrations/dockr"
	"os"
	"testing"
	"time"
	"golang.org/x/net/context"
)


// An extremely involved integration test with our local nexus repo and a docker image being brought up.
func TestDocker_RepoIntegrationSetup(t *testing.T) {
	// skip if is running in -short mode or if the nexus pw isn't set
	if testing.Short() {
		t.Skip("skipping due to -short flag being set")
	}
	password, ok := os.LookupEnv("NEXUS_ADMIN_PW")
	if !ok {
		t.Skip("skipping because $NEXUS_ADMIN_PW not set")
	}

	acctName := "test"
	projectName := "project"
	docker, cleanupFunc := CreateLivingDockerContainer(t, "docker:18.02.0-ce")
	defer cleanupFunc(t)

	pull := []string{"/bin/sh", "-c", "docker pull docker.metaverse.l11.com/busybox:test_do_not_delete"}
	su := InitStageUtil("testing")

	time.Sleep(time.Second)
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)
	defer cred.TeardownVaultAndConsul(vaultListener, consulServer)

	// try to do a docker pull on private repo without creds written. this should fail.
	out := make(chan []byte, 10000)
	ctx := context.Background()
	result := docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, out)
	if result.Status != pb.StageResultVal_FAIL {
		data := <-out
		t.Error("pull from metaverse should fail if there are no creds to authenticate with. stdout: " , data)
	}
	// add in docker repo credentials,
	cred.AddDockerRepoCreds(t, testRemoteConfig, "docker.metaverse.l11.com", password, "admin", acctName, projectName)

	// create config in ~/.docker directory w/ auth creds
	logout := make(chan[]byte, 10000)
	res := docker.IntegrationSetup(ctx, dockr.GetDockerConfig, docker.WriteDockerJson, "docker login", testRemoteConfig, acctName, su, []string{}, logout)
	if res.Status == pb.StageResultVal_FAIL {
		data := <- logout
		t.Error("stage failed! logout data: ", string(data))
	}

	// try to pull the image from metaverse again. this should now pass.
	logout = make(chan[]byte, 100000)
	res = docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, logout)
	outByte := <- logout
	if res.Status == pb.StageResultVal_FAIL {
		t.Error("could not pull from metaverse docker! out: ", string(outByte))
	}
}