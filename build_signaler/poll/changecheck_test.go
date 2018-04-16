package poll

import (
	"testing"

	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/common/remote"
	pbb "bitbucket.org/level11consulting/ocelot/models/bitbucket/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
)

type fakeCommitLister struct {
	commits []*pbb.Commit
	remote.VCSHandler
}

func (f *fakeCommitLister) GetAllCommits(string, string) (*pbb.Commits, error) {
	return &pbb.Commits{Values:f.commits}, nil
}

// faked all this out and wrote an interface because i only wanted to test the logic of whether or not this should trigger a build
type fakeWerkerTeller struct {}

func (f *fakeWerkerTeller) tellWerker(lastCommit *pbb.Commit, conf *ChangeChecker, branch string, store storage.OcelotStorage, remote remote.VCSHandler, token string) (err error) { return nil }

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
			commitList := []*pbb.Commit{{Hash:testcase.commitListHash}}
			commitListen := &fakeCommitLister{commits:commitList}
			store := storage.NewPostgresStorage("", "", "", 0, "")
			conf := &ChangeChecker{AcctRepo: "test/test"}
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