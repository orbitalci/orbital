package github

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)


func getCliAuth(t *testing.T) models.VCSHandler {
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	if token == "" {
		t.Skip("unit test requires authentication, and environment variable GITHUB_ACCESS_TOKEN not provided, skipping...")
	}
	creds := &pb.VCSCreds{
		ClientSecret: token,
	}
	cli, _, err := GetGithubClient(creds)
	if err != nil {
		t.Fatal(err)
	}
	return cli
}

func getCli(t *testing.T) models.VCSHandler {
	client := &net.OAuthClient{}
	client.SetupNoAuthentication()
	gh := GetGithubHandler(nil, client)
	return gh
}

func TestGithub_GetFile(t *testing.T) {
	cli := getCli(t)
	fil, err := cli.GetFile("README.md", "shankj3/test-ocelot", "74302362d6101d8675dfd6a99af7fec0b660ff94")
	if err != nil {
		t.Fatal(err)
	}
	live := string(fil)
	expected := `# test-ocelot
test repo for ocelot
`
	if !strings.EqualFold(expected, string(fil)) {
		t.Error(test.StrFormatErrors("readme contents", expected, live))
	}
}

func TestGithubVCS_GetRepoLinks(t *testing.T) {
	cli := getCli(t)
	links, err := cli.GetRepoLinks("shankj3/test-ocelot")
	if err != nil {
		t.Fatal(err)
	}
	linkz := &pb.Links{
		Commits: "https://api.github.com/repos/shankj3/test-ocelot/commits{/sha}",
		Branches: "https://api.github.com/repos/shankj3/test-ocelot/branches{/branch}",
		Tags: "https://api.github.com/repos/shankj3/test-ocelot/tags",
		Hooks: "https://api.github.com/repos/shankj3/test-ocelot/hooks",
		Pullrequests: "https://api.github.com/repos/shankj3/test-ocelot/pulls{/number}",
	}
	if diff := deep.Equal(links, linkz); diff != nil {
		t.Error(diff)
	}
}

func TestGithubVCS_CreateWebhook(t *testing.T) {
	cli := getCliAuth(t)
	hookUrl := "https://api.github.com/repos/shankj3/test-ocelot/hooks"
	if err := cli.CreateWebhook(hookUrl); err != nil {
		t.Error(err)
	}
}

func TestGithubVCS_GetAllBranchesLastCommitData(t *testing.T) {
	cli := getCli(t)
	allbranches, err := cli.GetAllBranchesLastCommitData("shankj3/test-ocelot")
	if err != nil {
		t.Fatal(err)
	}
	// sanity check that master branch exists
	for _, branch := range allbranches {
		if branch.Branch == "master" {
			return
		}
	}
	t.Error("branch master not in allbranches list")
}

func TestGithubVCS_GetBranchLastCommitData(t *testing.T) {
	cli := getCli(t)
	testBranch, err := cli.GetBranchLastCommitData("shankj3/test-ocelot", "fix/test-branch")
	if err != nil {
		t.Fatal(err)
		return
	}
	testBranchHeadSha := "094f6fc749227678a364f16a9d0f4ec3c41fdc8b"
	if testBranch.Branch != "fix/test-branch" {
		t.Error(test.StrFormatErrors("returned branch", "testBranch", testBranch.Branch))
	}
	if testBranch.Hash != testBranchHeadSha {
		t.Error(test.StrFormatErrors("returned last hash", testBranchHeadSha, testBranch.Hash))
	}
	commitTime := time.Unix(testBranch.LastCommitTime.Seconds, int64(testBranch.LastCommitTime.Nanos))
	expectedCommitTime, _ := time.Parse(time.UnixDate, "Tue Oct 23 09:48:37 PDT 2018")
	if !expectedCommitTime.Equal(commitTime) {
		t.Errorf("has someone updated the fix/test-branch branch? difference in commit times: \nexpected: %s live: %s", expectedCommitTime.Format(time.UnixDate), commitTime.Format(time.UnixDate))
	}
}

func TestGithubVCS_GetCommitLog(t *testing.T) {
	cli := getCli(t)
	lastHash := "f21cd5a1a71b106f1578356dc9d80fa174e23f69"
	log, err := cli.GetCommitLog("shankj3/test-ocelot", "fix/test-branch", lastHash)
	if err != nil {
		t.Fatal(err)
	}
	if len(log) != 8 {
		t.Fatal("should not return more than 8 commits")
	}
	if log[7].Hash != lastHash {
		t.Fatal("last commit should be the commit passed to GetCommitLog")
	}
}

func TestGithubVCS_PostPRComment(t *testing.T) {
	cli := getCliAuth(t)
	ghCli := cli.(*githubVCS)
	lastHash := "f21cd5a1a71b106f1578356dc9d80fa174e23f69"
	if err := cli.PostPRComment("shankj3/test-ocelot", "1", lastHash, true, 12); err != nil  {
		t.Fatal(err)
	}
	defer ghCli.deleteIssueComment("shankj3", "test-ocelot", ghCli.setCommentId)
	if err := ghCli.getIssueComment("shankj3", "test-ocelot", ghCli.setCommentId); err != nil {
		t.Fatal(err)
	}

}

func TestGithubVCS_GetChangedFiles(t *testing.T) {
	cli := getCli(t)
	latest := "094f6fc749227678a364f16a9d0f4ec3c41fdc8b"
	earliest := "b70d8b6d7fdee0886558142688c211464f84b20e"
	changedFiles, err := cli.GetChangedFiles("shankj3/test-ocelot", latest, earliest)
	if err != nil {
		t.Fatal(err)
	}
	expectedChanged := []string{".hi", "README.md"}
	if diff := deep.Equal(expectedChanged, changedFiles); diff != nil {
		t.Fatal(diff)
	}
}