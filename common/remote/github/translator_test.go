package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/models/pb"
)

var testRepo = &pb.Repo{
	Name:"test-ocelot",
	AcctRepo: "shankj3/test-ocelot",
	RepoLink: "https://api.github.com/repos/shankj3/test-ocelot",
}

func getWd(t *testing.T) string {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return pwd
}

func TestTranslator_TranslatePR(t *testing.T) {
	pwd := getWd(t)
	file, err := os.Open(filepath.Join(pwd, "test-fixtures", "github_pullrequest_opened.json"))
	if err != nil {
		t.Fatal(err)
	}
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
	pwd := getWd(t)
	file, err := os.Open(filepath.Join(pwd, "test-fixtures", "github_pullrequest_update.json"))
	if err != nil {
		t.Fatal(err)
	}
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