package build_signaler

import (
	"strings"
	//"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/mock_models"

	//"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
)

func TestPushWerkerTeller_TellWerker(t *testing.T) {
	cwt := &PushWerkerTeller{}
	signaler := GetFakeSignaler(t, false)
	handler := &DummyVcsHandler{Filecontents: Buildfile, Fail: false}
	var commits []*pb.Commit
	push := &pb.Push{
		HeadCommit: &pb.Commit{Hash: "hash"},
		Branch: "branch",
		Commits: commits,
		Repo: &pb.Repo{AcctRepo: "shankj3/ocelot"},
	}
	handler.ChangedFiles = []string{"changedfile.conf"}
	handler.ReturnCommit = &pb.Commit{Hash: "1234", Message: "this is my message"}
	err := cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err != nil {
		t.Error(err)
	}
	handler.Reset()
	err = cwt.TellWerker(push, signaler, handler, "", false, pb.SignaledBy_REQUESTED)
	if err == nil || err.Error() != "token cannot be empty" {
		t.Error("if token is emtpy, should return an error that says token cannot be empty")
	}
	queuedmsg := signaler.Producer.(*TestSingleProducer).Message.(*pb.WerkerTask)
	if queuedmsg.Id != 12 {
		t.Error(test.GenericStrFormatErrors("id", 12, queuedmsg.Id))
	}
	if queuedmsg.VaultToken != "token" {
		t.Error(test.StrFormatErrors("vault token", "token", queuedmsg.VaultToken))
	}
	handler.NotFound = true
	err = cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err == nil || err.Error() != "no ocelot yaml found for repo shankj3/ocelot" {
		t.Error("handler returned a file not found, shouldreturn an herror that ocelot yml can't be found")
	}
	handler.Reset()
	handler.Fail = true
	err = cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err == nil {
		t.Error("handler returned generic error, should be bubbled up to TellWerker caller")
	}
	signaler = GetFakeSignaler(t, true)
	handler = &DummyVcsHandler{Filecontents: Buildfile, Fail: false, ReturnCommit: &pb.Commit{Hash: "hash", Message: "commit msg"}}
	err = cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err == nil || !strings.Contains(err.Error(), "did not queue because it shouldn't be queued") {
		t.Error("if build is in consul, then should return a not viable error and shoudl bubble up to caller")
	}
}


func TestPushWerkerTeller_TellWerker_PreviousHeadCommit(t *testing.T) {
	ctl := gomock.NewController(t)
	handler := mock_models.NewMockVCSHandler(ctl)
	signaler := GetFakeSignaler(t, false)
	cwt := &PushWerkerTeller{}
	commits := []*pb.Commit{
		{Hash: "123last", Message: "mymessage is greaat", Author: &pb.User{UserName:"jessi-shank", DisplayName: "jessi shank"}},
		{Hash: "123secondLast", Message: "changing stuff its fine", Author: &pb.User{UserName:"jessi-shank", DisplayName: "jessi shank"}},
		{Hash: "123thirdlat", Message: "changing up some config", Author: &pb.User{UserName:"jessi-shank", DisplayName: "jessi shank"}},
		{Hash: "first", Message: "touching some new files!", Author: &pb.User{UserName:"jessi-shank", DisplayName: "jessi shank"}},
	}
	push := &pb.Push{
		Commits: commits,
		Repo: &pb.Repo{AcctRepo: "shankj3/ocelot"},
		HeadCommit: commits[0],
		PreviousHeadCommit: &pb.Commit{Hash: "old_last", Message: "finished that last pr!", Author: &pb.User{UserName:"jessi-shank", DisplayName: "jessi shank"}},
		User: &pb.User{UserName:"jessi-shank", DisplayName: "jessi shank"},
		Branch: "mybranchfornewstuffveryexcitingwoooooeeeeee",
	}
	changedFiles := []string{"ocelot.yml", "build.conf", "src/main/java/javathing.java"}
	handler.EXPECT().GetFile("ocelot.yml", "shankj3/ocelot", "123last").Times(1).Return(Buildfile, nil)
	handler.EXPECT().GetChangedFiles("shankj3/ocelot", "123last", "old_last").Times(1).Return(changedFiles, nil)
	if err := cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_PUSH); err != nil {
		t.Error("should not fail, got " + err.Error())
	}
}