package poll

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/models"
	pbb "github.com/shankj3/ocelot/models/bitbucket/pb"
	"github.com/shankj3/ocelot/models/pb"
)

type fakeCommitLister struct {
	commits []*pbb.Commit
	models.VCSHandler
	allBranchData []*pb.BranchHistory
}


func (f *fakeCommitLister) GetAllCommits(string, string) (*pbb.Commits, error) {
	return &pbb.Commits{Values: f.commits}, nil
}

func (f *fakeCommitLister) GetBranchLastCommitData(acctRepo, branch string) (*pb.BranchHistory, error) {
	return &pb.BranchHistory{Branch:branch, Hash:f.commits[0].Hash, LastCommitTime:&timestamp.Timestamp{}}, nil
}

func (f *fakeCommitLister) GetAllBranchesLastCommitData(acctRepo string) ([]*pb.BranchHistory, error) {
	if f.allBranchData == nil {
		return nil, errors.New("all branch data can't be nil")
	}
	return f.allBranchData, nil
}

func (f *fakeCommitLister) GetFile(string, string, string) ([]byte, error) {
	return []byte{}, nil
}

// faked all this out and wrote an interface because i only wanted to test the logic of whether or not this should trigger a build
type fakeWerkerTeller struct{
	told int
}

func (f *fakeWerkerTeller) TellWerker(lastCommit string, conf *build_signaler.Signaler, branch string, remote models.VCSHandler, token string) (err error) {
	f.told += 1
	return nil
}

var branchTests = []struct {
	name           string
	oldhash        string
	commitListHash string
	newHash        string
}{
	{"new commit", "boogaloo", "boogaboo", "boogaboo"},
	{"old commit", "oldie", "oldie", "oldie"},
	{"no last commit", "", "triggerMePlease", "triggerMePlease"},
}


func TestChangeChecker_InspectCommits(t *testing.T) {
	for _, testcase := range branchTests {
		t.Run(testcase.name, func(t *testing.T) {
			commitList := []*pbb.Commit{{Hash: testcase.commitListHash}}
			commitListen := &fakeCommitLister{commits: commitList}
			conf := &ChangeChecker{Signaler: &build_signaler.Signaler{AcctRepo: "test/test"}, handler:commitListen, token: "TOLKEIN", teller: &fakeWerkerTeller{}}
			//InspectCommits(branch string, lastHash string) (newLastHash string, err error) {
			newLastHash, err := conf.InspectCommits("test", testcase.oldhash)
			if err != nil {
				t.Error(err)
			}
			if newLastHash != testcase.newHash {
				t.Error(test.StrFormatErrors("returned hash", testcase.newHash, newLastHash))
			}

		})
	}
}

var allBranchesTests = []struct{
	name string
	histories []*pb.BranchHistory
	buildHashMap map[string]string
	finalBuildHashMap map[string]string
	toldTimes int
}{
	{
		"multiple branches should build",
		[]*pb.BranchHistory{{Hash:"1234", Branch:"abranch", LastCommitTime: &timestamp.Timestamp{Seconds:time.Now().Unix(), Nanos:0}},{Hash:"1a34", Branch:"bbranch", LastCommitTime: &timestamp.Timestamp{Seconds:time.Now().Unix(), Nanos:0}},},
		map[string]string{"abranch": "abcd", "bbranch": "12er"},
		map[string]string{"abranch": "1234", "bbranch": "1a34"},
		2,
	},
	{
		"no branches should build",
		[]*pb.BranchHistory{{Hash:"1234", Branch:"abranch", LastCommitTime: &timestamp.Timestamp{Seconds:time.Now().Unix(), Nanos:0}},{Hash:"1a34", Branch:"bbranch", LastCommitTime: &timestamp.Timestamp{Seconds:time.Now().Unix(), Nanos:0}},},
		map[string]string{"abranch":"1234", "bbranch": "1a34"},
		map[string]string{"abranch":"1234", "bbranch": "1a34"},
		0,
	},
	{
		"untracked branch",
		[]*pb.BranchHistory{{Hash:"1234", Branch:"abranch", LastCommitTime: &timestamp.Timestamp{Seconds:time.Now().Unix(), Nanos:0}},{Hash:"1a34", Branch:"bbranch", LastCommitTime: &timestamp.Timestamp{Seconds:time.Now().AddDate(0,0,6).Unix(), Nanos:0}},},
		map[string]string{"abranch":"1234"},
		map[string]string{"abranch":"1234", "bbranch": "1a34"},
		1,
	},
}

func TestChangeChecker_HandleAllBranches(t *testing.T) {
	for _, testcase := range allBranchesTests {
		t.Run(testcase.name, func(t *testing.T) {
			commitListen := &fakeCommitLister{allBranchData: testcase.histories}
			teller := &fakeWerkerTeller{}
			conf := &ChangeChecker{Signaler: &build_signaler.Signaler{AcctRepo: "test/test"}, handler:commitListen, token: "TOLKEIN", teller: teller}
			//InspectCommits(branch string, lastHash string) (newLastHash string, err error) {
			err := conf.HandleAllBranches(testcase.buildHashMap)
			if err != nil {
				t.Error(err)
			}
			if teller.told != testcase.toldTimes {
				t.Error(fmt.Sprintf("worker should have been told %d times, it was told %d", testcase.toldTimes, teller.told))
			}
			if diff := deep.Equal(testcase.buildHashMap, testcase.finalBuildHashMap); diff != nil {
				t.Error(diff)
			}
		})
	}
}
