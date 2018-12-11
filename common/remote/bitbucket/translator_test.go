package bitbucket

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/models/pb"
)

func TestBBTranslate_TranslatePR(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	file, err := os.Open(filepath.Join(pwd, "test-fixtures", "1528734656-pr-bb.json"))
	defer file.Close()
	trans := GetTranslator()
	pull, err := trans.TranslatePR(file)
	if err != nil {
		t.Error(err)
		return
	}
	if pull.Source.Hash != "dc128f78cd34" {
		t.Error(test.StrFormatErrors("source hash", "dc128f78cd34", pull.Source.Hash))
	}
	if pull.Source.Branch != "newbranch" {
		t.Error(test.StrFormatErrors("source branch", "newbranch", pull.Source.Branch))
	}
	if pull.Id != 1 {
		t.Error(test.GenericStrFormatErrors("pr id", 1, pull.Id))
	}
	commits := "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/1/commits"
	if pull.Urls.Commits != commits {
		t.Error(test.StrFormatErrors("commits url", commits, pull.Urls.Commits))
	}
	comments := "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/1/comments"
	if pull.Urls.Comments != comments {
		t.Error(test.StrFormatErrors("comments url", comments, pull.Urls.Comments))
	}
	if pull.Title != "Newbranch" {
		t.Error(test.StrFormatErrors("titile", "Newbranch", pull.Title))
	}
	if pull.Destination.Branch != "master" {
		t.Error(test.StrFormatErrors("dest branch", "master", pull.Destination.Branch))
	}
	if pull.Destination.Hash != "32ed49560d10" {
		t.Error(test.StrFormatErrors("dest hash", "32ed49560d10", pull.Destination.Hash))
	}
}

var jessiUser = &pb.User{DisplayName: "Jessi Shank", UserName: "jessi_shank"}
var jessiOwner = &pb.User{DisplayName: "Jessi Shank", UserName: "jessishank"}
var mytestOcy = &pb.Repo{
	Name: "jessishank/mytestocy",
	RepoLink: "https://bitbucket.org/jessishank/mytestocy",
	AcctRepo: "jessishank/mytestocy",
}

var pushTests = []struct{
	filename   string
	translated *pb.Push
	errorMsg   string
}{
	{
		filename: "push_new_branch_one_commit.json",
		translated: &pb.Push{
			Repo: mytestOcy,
			User: jessiOwner,
			HeadCommit: &pb.Commit{
				Hash: "c0984feed15b30e024016f47850402887c011a9e",
				Message: "add new file\n",
				Author: jessiUser,
				Date: &timestamp.Timestamp{Seconds:1538068054},
			},
			PreviousHeadCommit: &pb.Commit{
				Hash: "7e4a3761b5a5945367d62fb7f4acd374a0c18dd7",
				Message: "just going to do one commit this time\n",
				Author: jessiUser,
				Date: &timestamp.Timestamp{Seconds:1537997943},
			},
			Branch: "new_branch",
			Commits: []*pb.Commit{
				{
					Hash: "c0984feed15b30e024016f47850402887c011a9e",
					Message: "add new file\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1538068054},
				},
				// todo: a single push of a new branch will return the previous four commits too? idk why. but maybe we should
				// have different behavior.
				{
					Hash: "4f59e57d8878daabcbc6af6bd374929b9fe03568",
					Message: "adding test data\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1537998853},
				},
				{
					Hash: "48bd5377046bdd5bcce847c50b3bce90e22bd66d",
					Message: "another one!\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1537998814},
				},
				{
					Hash: "906a8969084b9ab0d8ce1056d3e046629eb2d6f9",
					Message: "adding one more change after submitting pr!!! woooahh nelly!\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1537998072},
				},
				{
					Hash: "7e4a3761b5a5945367d62fb7f4acd374a0c18dd7",
					Message: "just going to do one commit this time\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1537997943},
				},
			},
		},
	},
	{
		filename: "two_commit_push.json",
		translated: &pb.Push{
			Repo: mytestOcy,
			User: jessiOwner,
			Branch: "here_we_go",
			HeadCommit: &pb.Commit{
				Hash: "4f59e57d8878daabcbc6af6bd374929b9fe03568",
				Message: "adding test data\n",
				Author: jessiUser,
				Date: &timestamp.Timestamp{Seconds:1537998853},
			},
			PreviousHeadCommit: &pb.Commit{
				Hash: "906a8969084b9ab0d8ce1056d3e046629eb2d6f9",
				Message: "adding one more change after submitting pr!!! woooahh nelly!\n",
				Author: jessiUser,
				Date: &timestamp.Timestamp{Seconds:1537998072},
			},
			Commits: []*pb.Commit{
				{
					Hash: "4f59e57d8878daabcbc6af6bd374929b9fe03568",
					Message: "adding test data\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1537998853},
				},
				{
					Hash: "48bd5377046bdd5bcce847c50b3bce90e22bd66d",
					Message: "another one!\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1537998814},
				},
			},
		},
	},
	{
		filename: "push_after_pr_push_event.json",
		translated: &pb.Push{
			Repo: mytestOcy,
			User: jessiOwner,
			Branch: "here_we_go",
			HeadCommit: &pb.Commit{
				Hash: "906a8969084b9ab0d8ce1056d3e046629eb2d6f9",
				Message: "adding one more change after submitting pr!!! woooahh nelly!\n",
				Author: jessiUser,
				Date: &timestamp.Timestamp{Seconds:1537998072},
			},
			PreviousHeadCommit: &pb.Commit{
				Hash: "7e4a3761b5a5945367d62fb7f4acd374a0c18dd7",
				Message: "just going to do one commit this time\n",
				Author: jessiUser,
				Date: &timestamp.Timestamp{Seconds:1537997943},
			},
			Commits: []*pb.Commit{
				{
					Hash: "906a8969084b9ab0d8ce1056d3e046629eb2d6f9",
					Message: "adding one more change after submitting pr!!! woooahh nelly!\n",
					Author: jessiUser,
					Date: &timestamp.Timestamp{Seconds:1537998072},
				},
			},
		},
	},
	{
		filename: "two_changesets.json",
		errorMsg: "too many changesets",
	},
	{
		filename: "push_tag.json",
		errorMsg: "unexpected push type",
	},
}

func TestBBTranslate_TranslatePush(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	for _, unitTest := range pushTests {
		t.Run(unitTest.filename, func(t *testing.T) {
			file, err := os.Open(filepath.Join(pwd, "test-fixtures", unitTest.filename))
			defer file.Close()
			if err != nil {
				t.Error(err)
				return
			}
			trans := GetTranslator()
			push, err := trans.TranslatePush(file)
			if err != nil {
				if unitTest.errorMsg == err.Error() {
					return
				}
				t.Error(err)
				return
			}
			if diff := deep.Equal(push, unitTest.translated); diff != nil {
				t.Error(diff)
			}
		})
	}

}
