package webhook

import (
	"strings"
	"testing"

	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/models/pb"
)

func TestGetPrWerkerTeller(t *testing.T) {
	prwt := GetPrWerkerTeller("1", "branch")
	if prwt.prId != "1" || prwt.destBranch != "branch" {
		t.Error("not prwt not rendered properly")
	}
}

func TestPRWerkerTeller_TellWerker(t *testing.T) {
	prwt := GetPrWerkerTeller("1", "master")
	sig := build_signaler.GetFakeSignaler(t, false)
	handler := &build_signaler.DummyVcsHandler{NotFound:true}
	err := prwt.TellWerker("hash", sig, "feature", handler, "token", "shankj3/ocelot", []*pb.Commit{}, false, pb.SignaledBy_PULL_REQUEST)
	if err == nil {
		t.Error("error should not be nil")
	}
	if err.Error() != "no ocelot yml found for repo shankj3/ocelot" {
		t.Error("should have bubbled up notfound error that vcshandler threw")
	}
	handler = &build_signaler.DummyVcsHandler{Fail: true}
	err = prwt.TellWerker("hash", sig, "feature", handler, "token", "shankj3/ocelot", []*pb.Commit{}, false, pb.SignaledBy_PULL_REQUEST)
	if err == nil {
		t.Error("error should not be nil")
	}
	if !strings.Contains(err.Error(), "unable to get build con") {
		t.Error("should have bubbled up generic error that vcshandler threw, instead threw " + err.Error())
	}
	sig = build_signaler.GetFakeSignaler(t, true)
	handler = &build_signaler.DummyVcsHandler{Filecontents:build_signaler.Buildfile}
	err = prwt.TellWerker("hash", sig, "feature", handler, "token", "shankj3/ocelot", []*pb.Commit{}, false, pb.SignaledBy_PULL_REQUEST)
	if err == nil {
		t.Error("should return not viable error ")
	}
	if _, ok := err.(*build.NotViable); !ok {
		t.Error("if build is in consul should return a not viable error and bubble it up to prwt caller")
	}
	sig = build_signaler.GetFakeSignaler(t, false)
	handler = &build_signaler.DummyVcsHandler{Filecontents:build_signaler.BuildFileMasterOnly}
	err = prwt.TellWerker("hash", sig, "feature", handler, "token", "shankj3/ocelot", []*pb.Commit{}, false, pb.SignaledBy_PULL_REQUEST)
	if err != nil {
		t.Error("the build file says master only for building branches, and this pr is being merged to master, therefore this should build and stfu. the error is: " + err.Error())
	}
	sig = build_signaler.GetFakeSignaler(t, false)
	handler = &build_signaler.DummyVcsHandler{Filecontents:build_signaler.BuildFileMasterOnly}
	prwt.destBranch = "feature_2"
	err = prwt.TellWerker("hash", sig, "feature", handler, "token", "shankj3/ocelot", []*pb.Commit{}, false, pb.SignaledBy_PULL_REQUEST)
	if err == nil {
		t.Error("the build file says master only for building branches, and this pr is being merged to feature_2,therefore this build should not run")
	}


}