package poll

import (
	"testing"

	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/build_signaler"
	"bitbucket.org/level11consulting/ocelot/models"
	pbb "bitbucket.org/level11consulting/ocelot/models/bitbucket/pb"
)

type fakeCommitLister struct {
	commits []*pbb.Commit
	models.VCSHandler
}

func (f *fakeCommitLister) GetAllCommits(string, string) (*pbb.Commits, error) {
	return &pbb.Commits{Values:f.commits}, nil
}

// faked all this out and wrote an interface because i only wanted to test the logic of whether or not this should trigger a build
type fakeWerkerTeller struct {}

func (f *fakeWerkerTeller) TellWerker(lastCommit string, conf *build_signaler.Signaler, branch string, remote models.VCSHandler, token string) (err error) { return nil }

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
			conf := &ChangeChecker{Signaler: &build_signaler.Signaler{AcctRepo: "test/test"}}
			newLastHash, err := searchBranchCommits(commitListen, "test", conf, testcase.oldhash, "", &fakeWerkerTeller{})
			if err != nil {
				t.Error(err)
			}
			if newLastHash != testcase.newHash {
				t.Error(test.StrFormatErrors("returned hash", testcase.newHash, newLastHash))
			}

		})
	}
}