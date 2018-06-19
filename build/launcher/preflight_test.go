package launcher

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/build/integrations/sshkey"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

func Test_downloadCodebase(t *testing.T) {
	bilder := &fakeBuilder{
		setEnvs: []string{},
		Basher: getTestBasher(t),
	}
	ctx := context.Background()
	task := &pb.WerkerTask{
		VcsType: pb.SubCredType_BITBUCKET,
		CheckoutHash: "123",
		VcsToken: "token",
		FullName: "shankj3/ocelot",
	}
	logout := make(chan []byte, 100)
	stage := build.InitStageUtil("test")
	result := downloadCodebase(ctx, task, bilder, stage,  logout)
	close(logout)
	var output string
	for i:= range logout {
		output += string(i) + "\n"
	}
	if result.Status != pb.StageResultVal_PASS {
		t.Error("should have passed, output is " + output)
	}
	bilder.failExecuteIntegration = true
	logout = make(chan []byte, 100)
	result = downloadCodebase(ctx, task, bilder, stage,  logout)
	if result.Status != pb.StageResultVal_FAIL {
		t.Error("builder returned a failure, this should also fail.")
	}
}

func TestLauncher_preFlight(t *testing.T) {
	lnchr, clean := getTestingLauncher(t)
	defer clean(t)
	bilder := &fakeBuilder{
		setEnvs: []string{},
		Basher: getTestBasher(t),
	}
	ctx := context.Background()
	time.Sleep(5*time.Second)
	id, err := lnchr.Store.AddSumStart("123", "shankj3", "ocelot", "branch")
	if err != nil {
		t.Error(err)
		return
	}
	task := &pb.WerkerTask{
		VcsType: pb.SubCredType_BITBUCKET,
		CheckoutHash: "123",
		VcsToken: "token",
		FullName: "shankj3/ocelot",
		Branch: "branch",
		Id: id,
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