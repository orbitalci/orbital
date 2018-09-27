package build_signaler

import (
	"strings"
	//"strings"
	"testing"

	"github.com/shankj3/go-til/test"
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

	err := cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err != nil {
		t.Error(err)
	}
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
	handler = &DummyVcsHandler{NotFound: true}
	err = cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err == nil || err.Error() != "no ocelot yaml found for repo shankj3/ocelot" {
		t.Error("handler returned a file not found, shouldreturn an herror that ocelot yml can't be found")
	}
	handler = &DummyVcsHandler{Fail: true}
	err = cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err == nil {
		t.Error("handler returned generic error, should be bubbled up to TellWerker caller")
	}
	signaler = GetFakeSignaler(t, true)
	handler = &DummyVcsHandler{Filecontents: Buildfile, Fail: false}
	err = cwt.TellWerker(push, signaler, handler, "token", false, pb.SignaledBy_REQUESTED)
	if err == nil || !strings.Contains(err.Error(), "did not queue because it shouldn't be queued") {
		t.Error("if build is in consul, then should return a not viable error and shoudl bubble up to caller")
	}

}
