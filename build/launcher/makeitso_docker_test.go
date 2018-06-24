package launcher

import (
	"bytes"
	"strings"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/builder/docker"
	"github.com/shankj3/ocelot/build/integrations/dockerconfig"
	"github.com/shankj3/ocelot/common"
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
	&pb.K8SCreds{AcctName: "acct1", K8SContents:`herearemyk8scontentshowsickisthat
wowowoowowowoowoowowowo
herebeanotherlinebudshowniceisthaaat`, Identifier: "THERECANONLYBEONE", SubType: pb.SubCredType_KUBECONF},
&pb.K8SCreds{AcctName: "acct1", K8SContents:`Worst case Ontario`, Identifier: "ricky", SubType: pb.SubCredType_KUBECONF},
&pb.K8SCreds{AcctName: "acct1", K8SContents:`Propane
	and propane accessories
	I tell you hwhat`, Identifier: "HankHill", SubType: pb.SubCredType_KUBECONF},
}


var dockerCreds = []pb.OcyCredder{
	makeDockerCred("mydockeridentity", "dockeruser", "dockerpw", "http://urls.go", "jessdanshnak"),
	makeDockerCred("dockerid123", "uname", "xyzzzz", "http://docker.hub.io", "level11"),
}

var sshCreds = []pb.OcyCredder{
	&pb.SSHKeyWrapper{AcctName:"account1", PrivateKey:[]byte(testSSHKey), SubType:pb.SubCredType_SSHKEY, Identifier: "THISISANSSHKEY"},
}
var testSSHKey = `-----BEGIN RSA PRIVATE KEY-----
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww
-----END RSA PRIVATE KEY-----`

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
	case pb.SubCredType_SSHKEY:
		return sshCreds, nil
	}
	return nil, common.NCErr("only did docker and kubeconf")
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
	dckr, cleanupFunc := docker.CreateLivingDockerContainer(t, ctx, "docker:18.02.0-ce")
	time.Sleep(2 * time.Second)
	defer cleanupFunc(t)
	baseStage := build.InitStageUtil("PREFLIGHT")
	launch.infochan = make(chan []byte, 1000)
	launch.integrations = getIntegrationList()
	_ = dckr.Exec(ctx, "Install bash", "", []string{}, []string{"/bin/sh", "-c", "apk -U --no-cache add bash"}, launch.infochan) // For the k8s test
	result := launch.doIntegrations(ctx, &pb.WerkerTask{BuildConf: &pb.BuildConfig{BuildTool: "maven"}}, dckr, baseStage)
	if result.Status == pb.StageResultVal_FAIL {
		t.Log(result.Messages)
		t.Error(result.Error)
	}
	expectedMsgs := []string{
		"completed preflight | integ | ssh keyfile integration stage ✓",
		"completed preflight | integ | docker login stage ✓",
		"completed preflight | integ | kubeconfig render stage ✓",
		"no integration data for nexus m2 settings.xml render ✓",
		"completed integration util setup stage ✓",
	}
	if diff := deep.Equal(expectedMsgs, result.Messages); diff != nil {
		t.Error(diff)
	}
	result = launch.doIntegrations(ctx, &pb.WerkerTask{BuildConf:&pb.BuildConfig{BuildTool:"gala"}}, dckr, baseStage)
	close(launch.infochan)
	expectedMsgs = []string{
		"completed preflight | integ | ssh keyfile integration stage ✓",
		"completed preflight | integ | docker login stage ✓",
		"completed preflight | integ | kubeconfig render stage ✓",
		"completed integration util setup stage ✓",
	}
	if diff := deep.Equal(expectedMsgs, result.Messages); diff != nil {
		var i []byte
		i =  <- launch.infochan
		t.Log(string(i))
		//t.Log(<-launch.infochan)
		t.Error(diff)
	}
	launch.infochan = make(chan []byte, 1000)
	// check that docker config.json was properly rendered
	expectedDockerConfig := []byte(`{"auths":{"http://docker.hub.io":{"auth":"dW5hbWU6eHl6enp6"},"http://urls.go":{"auth":"ZG9ja2VydXNlcjpkb2NrZXJwdw=="}},"HttpHeaders":{"User-Agent":"Docker-Client/17.12.0-ce (linux)"}}`)
	res := dckr.Exec(ctx, "test docker config", "", []string{}, []string{"/bin/sh", "-c", "cat ~/.docker/config.json"}, launch.infochan)
	if res.Status == pb.StageResultVal_FAIL {
		t.Log(res.Messages)
		t.Error(result.Error)
	}
	config := <- launch.infochan
	if !bytes.Equal(expectedDockerConfig, config) {
		t.Error(test.GenericStrFormatErrors("docker config contents", string(expectedDockerConfig), string(config)))
	}
	close(launch.infochan)

	// check that kubeconfig was properly rendered

	// this is going to be multiline, so we have to set up a new info channel
	// kubeconf[0]'s identifier is THERECANONLYBEONE, which we expect to be written to ~/.kube/config
	kubeLogOne := make(chan []byte, 1000)
	res = dckr.Exec(ctx, "test k8s config, identifier 'THERECANONLYBEONE'", "", []string{}, []string{"/bin/sh", "-c", "cat ~/.kube/config"}, kubeLogOne)
	close(kubeLogOne)
	if res.Status == pb.StageResultVal_FAIL {
		t.Log(res.Messages)
		t.Error(result.Error)
	}
	var kubelinesOne []string
	for kubeline := range kubeLogOne {
		kubelinesOne = append(kubelinesOne, string(kubeline))
	}
	kubeOne := strings.Join(kubelinesOne, "\n")
	if kubeOne != kubeconfs[0].GetClientSecret() {
		t.Error(test.StrFormatErrors("kubeconfig contents", kubeconfs[0].GetClientSecret(), kubeOne))
	}

	// Testing that a new, named k8s creds renders with a user provided identifier
	// kubeconf[1]'s identifier is ricky, which we expect to be written to ~/.kube/ricky
	kubeLogTwo := make(chan []byte, 1000)
	res = dckr.Exec(ctx, "test k8s config, identifier 'ricky'", "", []string{}, []string{"/bin/sh", "-c", "cat ~/.kube/ricky"}, kubeLogTwo)
	close(kubeLogTwo)
	if res.Status == pb.StageResultVal_FAIL {
		t.Log(res.Messages)
		t.Error(result.Error)
	}
	var kubelinesTwo []string
	for kubeline := range kubeLogTwo {
		kubelinesTwo = append(kubelinesTwo, string(kubeline))
	}
	kubeTwo := strings.Join(kubelinesTwo, "\n")
	if kubeTwo != kubeconfs[1].GetClientSecret() {
		t.Error(test.StrFormatErrors("kubeconfig contents", kubeconfs[1].GetClientSecret(), kubeTwo))
	}

	// kubeconf[2]'s identifier is HankHill, which we expect to be written to ~/.kube/HankHill
	kubeLogThree := make(chan []byte, 1000)
	res = dckr.Exec(ctx, "test k8s config, identifier 'HankHill'", "", []string{}, []string{"/bin/sh", "-c", "cat ~/.kube/HankHill"}, kubeLogThree)
	close(kubeLogThree)
	if res.Status == pb.StageResultVal_FAIL {
		t.Log(res.Messages)
		t.Error(result.Error)
	}
	var kubelinesThree []string
	for kubeline := range kubeLogThree {
		kubelinesThree = append(kubelinesThree, string(kubeline))
	}
	kubeThree := strings.Join(kubelinesThree, "\n")
	if kubeThree != kubeconfs[2].GetClientSecret() {
		t.Error(test.StrFormatErrors("kubeconfig contents", kubeconfs[2].GetClientSecret(), kubeThree))
	}

	// finally, check that ssh key properly rendered
	sshLogout := make(chan []byte, 1000)
	res = dckr.Exec(ctx, "test ssh file render", "", []string{}, []string{"/bin/sh", "-c", "cat ~/.ssh/THISISANSSHKEY"}, sshLogout)
	close(sshLogout)
	var sshKeyRendered string
	for sshLIne := range sshLogout {
		sshKeyRendered += string(sshLIne) + "\n"
	}
	// add one more newline to expected as a result of the script that echoed the env var into the container
	testSSHKey += "\n"
	if sshKeyRendered  != testSSHKey {
		t.Error(test.StrFormatErrors("ssh file contents", testSSHKey, sshKeyRendered))
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
	dckr, cleanupFunc := docker.CreateLivingDockerContainer(t, ctx, "docker:18.02.0-ce")
	defer cleanupFunc(t)

	pull := []string{"/bin/sh", "-c", "docker pull docker.metaverse.l11.com/busybox:test_do_not_delete"}
	su := build.InitStageUtil("testing")

	time.Sleep(time.Second)
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)
	defer cred.TeardownVaultAndConsul(vaultListener, consulServer)

	// try to do a docker pull on private repo without creds written. this should fail.
	out := make(chan []byte, 10000)
	result := dckr.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, out)
	if result.Status != pb.StageResultVal_FAIL {
		data := <-out
		t.Error("pull from metaverse should fail if there are no creds to authenticate with. stdout: ", data)
	}
	// add in docker repo credentials,
	cred.AddDockerRepoCreds(t, testRemoteConfig, pg, "docker.metaverse.l11.com", password, "admin", acctName, projectName)
	//
	//// create config in ~/.docker directory w/ auth creds
	logout := make(chan []byte, 10000)
	dcker := dockerconfig.Create()
	creds, err := testRemoteConfig.GetCredsBySubTypeAndAcct(pg, dcker.SubType(), acctName, false)
	if err != nil {
		t.Log(err)
		return
	}
	intstring, err := dcker.GenerateIntegrationString(creds)
	if err != nil {
		t.Log(err)
		return
	}
	res := dckr.ExecuteIntegration(ctx, &pb.Stage{Env: []string{}, Name: "docker login", Script: dcker.MakeBashable(intstring)}, build.InitStageUtil("docker login"), logout)
	if res.Status == pb.StageResultVal_FAIL {
		data := <-logout
		t.Error("stage failed! logout data: ", string(data))
	}

	// try to pull the image from metaverse again. this should now pass.
	logout = make(chan []byte, 100000)
	res = dckr.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, logout)
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
