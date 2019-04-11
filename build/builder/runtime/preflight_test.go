package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/build/integrations"
	"github.com/level11consulting/ocelot/build/integrations/sshkey"
	"github.com/level11consulting/ocelot/common/credentials"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/test"
)

func Test_downloadCodebase(t *testing.T) {
	bilder := &fakeBuilder{
		setEnvs: []string{},
		Basher:  getTestBasher(t),
	}
	ctx := context.Background()
	task := &pb.WerkerTask{
		VcsType:      pb.SubCredType_BITBUCKET,
		CheckoutHash: "123",
		VcsToken:     "token",
		FullName:     "level11consulting/ocelot",
	}
	logout := make(chan []byte, 100)
	stage := build.InitStageUtil("test")
	result := downloadCodebase(ctx, task, bilder, stage, logout)
	close(logout)
	var output string
	for i := range logout {
		output += string(i) + "\n"
	}
	if result.Status != pb.StageResultVal_PASS {
		t.Error("should have passed, output is " + output)
	}
	bilder.failExecuteIntegration = true
	logout = make(chan []byte, 100)
	result = downloadCodebase(ctx, task, bilder, stage, logout)
	if result.Status != pb.StageResultVal_FAIL {
		t.Error("builder returned a failure, this should also fail.")
	}
}

func TestLauncher_preFlight(t *testing.T) {
	t.Skip(t)
	lnchr, _ := getTestingLauncher(t)
	time.Sleep(3 * time.Second)
	//defer clean(t)
	bilder := &fakeBuilder{
		setEnvs: []string{},
		Basher:  getTestBasher(t),
	}
	ctx := context.Background()
	time.Sleep(5 * time.Second)
	id, err := lnchr.Store.AddSumStart("123", "shankj3", "ocelot", "branch")
	if err != nil {
		t.Error(err)
		return
	}
	task := &pb.WerkerTask{
		VcsType:      pb.SubCredType_BITBUCKET,
		CheckoutHash: "123",
		VcsToken:     "token",
		FullName:     "level11consulting/ocelot",
		Branch:       "branch",
		Id:           id,
	}
	_, err = lnchr.preFlight(ctx, task, bilder)
	if err != nil {
		t.Error(err)
	}
	var bailOut bool
	bilder.failExecuteIntegration = true
	// add ssh key creds so integrations block will run
	err = lnchr.RemoteConf.AddCreds(lnchr.Store, &pb.SSHKeyWrapper{AcctName: "shankj3", Identifier: "identity", SubType: pb.SubCredType_SSHKEY, PrivateKey: []byte("so private")}, true)
	if err != nil {
		t.Error(err)
	}
	lnchr.integrations = []integrations.StringIntegrator{sshkey.Create()}
	bailOut, _ = lnchr.preFlight(ctx, task, bilder)
	if !bailOut {
		t.Error("builder failed, preFlight should be tellingi ts caller its time to give up")
	}

	lnchr.integrations = []integrations.StringIntegrator{}
	lnchr.Store = &fakeStore{fail: true}
	_, err = lnchr.preFlight(ctx, task, bilder)
	if err == nil {
		t.Error("storage returned a faileure, this should fail")
	}
	if err.Error() != "i fail now, ok?" {
		t.Error(test.StrFormatErrors("err msg", "i fail now, ok?", err.Error()))
	}
}

func TestLauncher_handleEnvSecrets(t *testing.T) {
	creds := []pb.OcyCredder{
		&pb.GenericCreds{
			AcctName:     "oooooops",
			Identifier:   "noicenoice",
			ClientSecret: "thisissecret",
			SubType:      pb.SubCredType_ENV,
		},
		&pb.GenericCreds{
			AcctName:     "oooooops",
			Identifier:   "GIT_SECRET",
			ClientSecret: "mewmew",
			SubType:      pb.SubCredType_ENV,
		},
		&pb.GenericCreds{
			AcctName:     "oooooops",
			Identifier:   "ddd",
			ClientSecret: "showme==",
			SubType:      pb.SubCredType_ENV,
		},
	}
	rc := &remoteConf{creds: creds}
	lnchr := &launcher{RemoteConf: rc}
	bilder := &fakeBuilder{
		setEnvs:   []string{},
		addedEnvs: []string{},
		Basher:    getTestBasher(t),
	}
	res := lnchr.handleEnvSecrets(context.Background(), bilder, "oooooops", build.InitStageUtil("PREFLIGHT"))
	if res.Status == pb.StageResultVal_FAIL {
		t.Error(res.Error)
	}
	expectedEnvs := []string{
		"noicenoice=thisissecret",
		"GIT_SECRET=mewmew",
		"ddd=showme==",
	}
	if diff := deep.Equal(expectedEnvs, bilder.addedEnvs); diff != nil {
		t.Error(diff)
	}

}

type remoteConf struct {
	creds []pb.OcyCredder
	credentials.CVRemoteConfig
}

func (rc *remoteConf) GetCredsBySubTypeAndAcct(store storage.CredTable, stype pb.SubCredType, accountName string, hideSecret bool) ([]pb.OcyCredder, error) {
	return rc.creds, nil
}

type fakeStore struct {
	fail bool
	storage.OcelotStorage
}

func (f *fakeStore) AddStageDetail(stageResult *models.StageResult) error {
	if f.fail {
		return errors.New("i fail now, ok?")
	}
	return nil
}

func (f *fakeStore) RetrieveCredBySubTypeAndAcct(scredType pb.SubCredType, acctName string) ([]pb.OcyCredder, error) {
	return nil, nil
}
