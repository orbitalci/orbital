package launcher

import (
	"github.com/go-test/deep"
	"github.com/shankj3/go-til/net"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/builder"
	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/build/integrations/dockerconfig"
	cred "github.com/shankj3/ocelot/common/credentials"

	//"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"

	"context"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"testing"
	"time"
)

var kubeconfs = []pb.OcyCredder{
	&pb.K8SCreds{"acct1", "herearemyk8scontentshowsickisthat", "THERECANONLYBEONE", pb.SubCredType_KUBECONF},
}

var dockerCreds = []pb.OcyCredder{
	makeDockerCred("mydockeridentity", "dockeruser", "dockerpw", "http://urls.go", "jessdanshnak"),
	makeDockerCred("dockerid123", "uname", "xyzzzz", "http://docker.hub.io", "level11"),
}

func makeDockerCred(id, uname, pw, url, acctname string) *pb.RepoCreds {
	return &pb.RepoCreds{Identifier: id, Username: uname, Password: pw, RepoUrl: url, AcctName: acctname, SubType: pb.SubCredType_DOCKER}
}

type dummyCVRC struct {
	cred.CVRemoteConfig
}

func (d *dummyCVRC) GetCredsBySubTypeAndAcct(store storage.CredTable, stype pb.SubCredType, accountName string, hideSecret bool) ([]pb.OcyCredder, error) {
	switch stype {
	case pb.SubCredType_KUBECONF:
		return kubeconfs, nil
	case pb.SubCredType_DOCKER:
		return dockerCreds, nil
	}
	return nil, integrations.NCErr("only did docker and kubeconf")
}

func createKubectlEndpoint(t *testing.T) {
	muxi := mux.NewRouter()
	muxi.HandleFunc("/kubectl", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://storage.googleapis.com/kubernetes-release/release/v1.9.6/bin/linux/amd64/kubectl", 301)
	})

	n := net.InitNegroni("werker", muxi)
	n.Run(":8888")
}

func TestLauncher_doIntegrations(t *testing.T) {
	launch := &launcher{RemoteConf: &dummyCVRC{}}
	//go createKubectlEndpoint(t)
	ctx := context.Background()
	docker, cleanupFunc := builder.CreateLivingDockerContainer(t, ctx, "docker:18.02.0-ce")
	time.Sleep(2 * time.Second)
	defer cleanupFunc(t)
	launch.infochan = make(chan []byte, 1000)
	result, _, _ := launch.doIntegrations(ctx, &pb.WerkerTask{BuildConf: &pb.BuildConfig{BuildTool: "maven"}}, docker)
	if result.Status == pb.StageResultVal_FAIL {
		t.Log(result.Messages)
		t.Error(result.Error)
	}
	expectedMsgs := []string{
		"no integration data found for ssh keyfile integration so assuming integration not necessary",
		"completed integration_util | docker login stage ✓",
		"completed integration_util | kubeconfig render stage ✓",
		"no integration data found for nexus m2 settings.xml render so assuming integration not necessary",
		"completed integration util setup stage ✓",
	}
	if diff := deep.Equal(expectedMsgs, result.Messages); diff != nil {
		t.Error(diff)
	}
	result, _, _ = launch.doIntegrations(ctx, &pb.WerkerTask{BuildConf: &pb.BuildConfig{BuildTool: "gala"}}, docker)
	expectedMsgs = []string{
		"no integration data found for ssh keyfile integration so assuming integration not necessary",
		"completed integration_util | docker login stage ✓",
		"completed integration_util | kubeconfig render stage ✓",
		"completed integration util setup stage ✓",
	}
	if diff := deep.Equal(expectedMsgs, result.Messages); diff != nil {
		t.Error(diff)
	}

}

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
	cleanup, pw, port := storage.CreateTestPgDatabase(t)
	defer cleanup(t)
	pg := storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")

	acctName := "test"
	projectName := "project"
	ctx := context.Background()
	docker, cleanupFunc := builder.CreateLivingDockerContainer(t, ctx, "docker:18.02.0-ce")
	defer cleanupFunc(t)

	pull := []string{"/bin/sh", "-c", "docker pull docker.metaverse.l11.com/busybox:test_do_not_delete"}
	su := build.InitStageUtil("testing")

	time.Sleep(time.Second)
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)
	defer cred.TeardownVaultAndConsul(vaultListener, consulServer)

	// try to do a docker pull on private repo without creds written. this should fail.
	out := make(chan []byte, 10000)
	result := docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, out)
	if result.Status != pb.StageResultVal_FAIL {
		data := <-out
		t.Error("pull from metaverse should fail if there are no creds to authenticate with. stdout: ", data)
	}
	// add in docker repo credentials,
	cred.AddDockerRepoCreds(t, testRemoteConfig, pg, "docker.metaverse.l11.com", password, "admin", acctName, projectName)
	//
	//// create config in ~/.docker directory w/ auth creds
	logout := make(chan []byte, 10000)
	dckr := dockerconfig.Create()
	creds, err := testRemoteConfig.GetCredsBySubTypeAndAcct(pg, dckr.SubType(), acctName, false)
	if err != nil {
		t.Log(err)
		return
	}
	intstring, err := dckr.GenerateIntegrationString(creds)
	if err != nil {
		t.Log(err)
		return
	}
	res := docker.ExecuteIntegration(ctx, &pb.Stage{Env: []string{}, Name: "docker login", Script: dckr.MakeBashable(intstring)}, build.InitStageUtil("docker login"), logout)
	if res.Status == pb.StageResultVal_FAIL {
		data := <-logout
		t.Error("stage failed! logout data: ", string(data))
	}

	// try to pull the image from metaverse again. this should now pass.
	logout = make(chan []byte, 100000)
	res = docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, logout)
	outByte := <-logout
	if res.Status == pb.StageResultVal_FAIL {
		t.Error("could not pull from metaverse docker! out: ", string(outByte))
	}
	muxi := mux.NewRouter()
	muxi.HandleFunc("/kubectl", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://storage.googleapis.com/kubernetes-release/release/v1.9.6/bin/linux/amd64/kubectl", 301)
	})
	//
	//n := net.InitNegroni("werker", muxi)
	//go n.Run(":8888")

	//result = docker.IntegrationSetup(ctx, func(config cred.CVRemoteConfig, store storage.CredTable, acctName string)(string, error) {return "8888", nil}, docker.DownloadKubectl, "kubectl download", testRemoteConfig, acctName, su, []string{}, pg, logout)
	//outBytes := <-logout
	//if result.Status == pb.StageResultVal_FAIL {
	//	t.Error("couldn't download kubectl! oh nuuuuuu! ", string(outBytes))
	//}
	//checkKube := []string{"/bin/sh", "-c", "command -v kubectl"}
	//outd := make(chan []byte, 10000)
	//result = docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, checkKube, outd)
	//outb := <- outd
	//if result.Status == pb.StageResultVal_FAIL {
	//	t.Error("kubectl not found! fail! ", string(outb))
	//}
	//
	//result = docker.IntegrationSetup(ctx, func(config cred.CVRemoteConfig, store storage.CredTable,  acctName string)(string, error) {return "dGhpc2lzbXlrdWJlY29uZm9oaGhoYmJiYWFhYWJiYmJ5eXl5Cg==", nil}, docker.InstallKubeconfig, "kubeconfig install", testRemoteConfig, acctName, su, []string{}, pg, logout)
	//outBytes = <-logout
	//if result.Status == pb.StageResultVal_FAIL {
	//	t.Error("couldn't add kubeconfig! oh no! ", string(outBytes))
	//}
	//t.Log(result.Status)
	//t.Log(string(outb))
	//checkKubeConf := []string{"/bin/sh", "-c", "cat ~/.kube/config"}
	//outx := make(chan []byte, 10000)
	//result = docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, checkKubeConf, outx)
	//outy := <- outx
	//if result.Status == pb.StageResultVal_FAIL {
	//	t.Error("kubeconf when wrong! fail! ", string(outy))
	//}
	//if string(outy) != "TESTING | thisismykubeconfohhhhbbbaaaabbbbyyyy" {
	//	t.Error(test.StrFormatErrors("kubeconfig contents", "TESTING | thisismykubeconfohhhhbbbaaaabbbbyyyy", string(outy)))
	//}

}
