package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/timestamp"
	//"github.com/pkg/errors"
	"github.com/level11consulting/orbitalci/models/pb"
)

var testRepo = &pb.Repo{
	Name:"test-ocelot",
	AcctRepo: "shankj3/test-ocelot",
	RepoLink: "https://api.github.com/repos/shankj3/test-ocelot",
}

var testRepoNoApi = &pb.Repo{
	Name:"test-ocelot",
	AcctRepo: "shankj3/test-ocelot",
	RepoLink: "https://github.com/shankj3/test-ocelot",
}

var testUser = &pb.User{
	UserName: "shankj3",
	DisplayName: "Jessi Shank",
}

func getWd(t *testing.T) string {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return pwd
}

func getFile(t *testing.T, filename string) *os.File {
	pwd := getWd(t)
	file, err := os.Open(filepath.Join(pwd, "test-fixtures", filename))
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestTranslator_TranslatePR(t *testing.T) {
	file := getFile(t, "github_pullrequest_opened.json")
	defer file.Close()
	trans := &translator{}
	pr, err := trans.TranslatePR(file)
	if err != nil {
		t.Fatal(err)
	}
	expectedPr := &pb.PullRequest{
		Title: "Thisbetestboyyyy dj khaled",
		Urls: &pb.PrUrls{
			Commits: "https://api.github.com/repos/shankj3/test-ocelot/pulls/2/commits",
			Comments: "https://api.github.com/repos/shankj3/test-ocelot/issues/2/comments",
			Statuses: "https://api.github.com/repos/shankj3/test-ocelot/statuses/3d551291d03b27d52b2f42eaf8e7ccc8d7ca4a06",
		},
		Source: &pb.HeadData{
			Branch: "thisbetestboyyyy_dj_KHALED",
			Hash: "3d551291d03b27d52b2f42eaf8e7ccc8d7ca4a06",
			Repo: testRepo,
		},
		Destination: &pb.HeadData{
			Branch: "master",
			Hash: "74302362d6101d8675dfd6a99af7fec0b660ff94",
			Repo: testRepo,
		},
		Id: 2,
	}
	if diff := deep.Equal(expectedPr, pr); diff != nil {
		t.Error(diff)
	}
}

func TestTranslator_TranslatePR_update(t *testing.T) {
	file := getFile(t, "github_pullrequest_update.json")
	defer file.Close()
	trans := &translator{}
	pr, err := trans.TranslatePR(file)
	if err != nil {
		t.Fatal(err)
	}
	expectedPr := &pb.PullRequest{
		Title: "Thisbetestboyyyy dj khaled",
		Urls: &pb.PrUrls{
			Commits: "https://api.github.com/repos/shankj3/test-ocelot/pulls/2/commits",
			Comments: "https://api.github.com/repos/shankj3/test-ocelot/issues/2/comments",
			Statuses: "https://api.github.com/repos/shankj3/test-ocelot/statuses/1defca0566af9fb8c6f9a50e6fdc61820e6b1f13",
		},
		Source: &pb.HeadData{
			Branch: "thisbetestboyyyy_dj_KHALED",
			Hash: "1defca0566af9fb8c6f9a50e6fdc61820e6b1f13",
			Repo: testRepo,
		},
		Destination: &pb.HeadData{
			Branch: "master",
			Hash: "74302362d6101d8675dfd6a99af7fec0b660ff94",
			Repo: testRepo,
		},
		Id: 2,
	}
	if diff := deep.Equal(expectedPr, pr); diff != nil {
		t.Error(diff)
	}
}

func TestTranslator_TranslatePush_newbranch(t *testing.T) {
	file := getFile(t, "github_push_new_branch.json")
	defer file.Close()
	trans := &translator{}
	push, err := trans.TranslatePush(file)
	if err != nil {
		t.Fatal(err)
	}
	expected := &pb.Push{
		Repo: testRepoNoApi,
		User: &pb.User{UserName: "shankj3"},
		HeadCommit: &pb.Commit{
			Message: "newfile update",
			Hash: "e23fc2e3540bfabfffe0487fa24557e58ace1749",
			Date: &timestamp.Timestamp{Seconds: 1540316553},
			Author: testUser,
		},
		Branch: "thisbetestboyyyy_ANOTHERONE",
		PreviousHeadCommit: nil,
		Commits: nil,
	}
	if diff := deep.Equal(expected, push); diff != nil {
		t.Error(diff)
	}
}

func TestTranslator_TranslatePush_tag(t *testing.T) {
	file := getFile(t, "github_push_tag.json")
	defer file.Close()
	trans := &translator{}
	_, err := trans.TranslatePush(file)
	if err == nil {
		t.Error("should error out")
	}
	if err.Error() != "unexpected push type" {
		t.Error("should be bad push type, got" + err.Error())
	}
}

func TestTranslator_TranslatePush_new_branch_commits(t *testing.T) {
	file := getFile(t,"github_push_new_branch_with_commits.json")
	defer file.Close()
	trans := &translator{}
	push, err := trans.TranslatePush(file)
	if err != nil {
		t.Error(err)
	}
	expected := &pb.Push{
		Repo: testRepoNoApi,
		Branch: "thisbetestboyyyy_dj_KHALED",
		User: &pb.User{UserName: "shankj3"},
		HeadCommit: &pb.Commit{
			Hash: "3d551291d03b27d52b2f42eaf8e7ccc8d7ca4a06",
			Message: "adding to newfile",
			Date: &timestamp.Timestamp{Seconds: 1540316861},
			Author: testUser,
		},
		PreviousHeadCommit: &pb.Commit{
			Hash: "408791f359227a7c48a1a0acdaff6528f634b4dc",
			Message: "adding to .hi",
			Date: &timestamp.Timestamp{Seconds:1540316835},
			Author: testUser,
		},
		Commits: []*pb.Commit{
			{
				Hash: "408791f359227a7c48a1a0acdaff6528f634b4dc",
				Message: "adding to .hi",
				Date: &timestamp.Timestamp{Seconds:1540316835},
				Author: testUser,
			},
			{
				Hash: "3d551291d03b27d52b2f42eaf8e7ccc8d7ca4a06",
				Message: "adding to newfile",
				Date: &timestamp.Timestamp{Seconds: 1540316861},
				Author: testUser,
			},
		},
	}
	if diff := deep.Equal(expected, push); diff != nil {
		t.Error(diff)
	}
}

func TestTranslator_TranslatePush_github_push_one_commit(t *testing.T) {
	file := getFile(t, "github_push_one_commit.json")
	defer file.Close()
	trans := &translator{}
	push, err := trans.TranslatePush(file)
	if err != nil {
		t.Error(err)
	}
	expected := &pb.Push{
		Repo: testRepoNoApi,
		Branch: "thisbetestboyyyy",
		User: &pb.User{UserName: "shankj3"},
		HeadCommit: &pb.Commit{
			Hash: "e23fc2e3540bfabfffe0487fa24557e58ace1749",
			Message: "newfile update",
			Date: &timestamp.Timestamp{Seconds: 1540316553},
			Author: testUser,
		},
		PreviousHeadCommit: &pb.Commit{
			Hash: "79062bb7109556f0820068a9dd9b9886564a2e8b",
		},
		Commits: []*pb.Commit{
			{
				Hash: "e23fc2e3540bfabfffe0487fa24557e58ace1749",
				Message: "newfile update",
				Date: &timestamp.Timestamp{Seconds: 1540316553},
				Author: testUser,
			},
		},
	}
	if diff := deep.Equal(expected, push); diff != nil {
		t.Error(diff)
	}

}
