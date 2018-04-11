package main

import (
	"testing"

	"bitbucket.org/level11consulting/go-til/test"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
)

type fakeCommitLister struct {
	commits []*pb.Commit
	handler.VCSHandler
}

func (f *fakeCommitLister) GetAllCommits(string, string) (*pb.Commits, error) {
	return &pb.Commits{Values:f.commits}, nil
}

// faked all this out and wrote an interface because i only wanted to test the logic of whether or not this should trigger a build
type fakeWerkerTeller struct {}

func (f *fakeWerkerTeller) tellWerker(lastCommit *pb.Commit, conf *changeSetConfig, branch string, store storage.OcelotStorage, handler handler.VCSHandler, token string) (err error) { return nil }

var branchTests = []struct {
	name		    string
	oldhash 	    string
	commitListHash  string
	newHash         string
}{
	{"new commit", "boogaloo", "boogaboo", "boogaboo"},
	{"old commit", "oldie", "oldie", "oldie"},
    {"no last commit", "", "triggerMePlease", "triggerMePlease"},
}

func Test_searchBranchCommits(t *testing.T) {
	for _, testcase := range branchTests {
		t.Run(testcase.name, func(t *testing.T){
			commitList := []*pb.Commit{{Hash:testcase.commitListHash}}
			commitListen := &fakeCommitLister{commits:commitList}
			store := storage.NewPostgresStorage("", "", "", 0, "")
			conf := &changeSetConfig{AcctRepo: "test/test"}
			newLastHash, err := searchBranchCommits(commitListen, "test", conf, testcase.oldhash, store, "", &fakeWerkerTeller{})
			if err != nil {
				t.Error(err)
			}
			if newLastHash != testcase.newHash {
				t.Error(test.StrFormatErrors("returned hash", testcase.newHash, newLastHash))
			}

		})
	}
}