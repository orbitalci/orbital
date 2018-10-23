package github

import (
	"strings"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-github/github"
	"github.com/shankj3/ocelot/models/pb"
)

var (
	REPOS = "repos"
	FILE = REPOS + "/%s/contents/%s"
)

// url replacements for githubVCS, see below for urls returned and what to replace

var (
	//"contents_url": "https://api.github.com/repos/shankj3/legis_data/contents/{+path}",
	CONTENTS_URL_REPLACE = "{+path}"
)

func getUrlForFileFromContentsUrl(contentsUrl string, relativeFilepath string) string {
	return strings.Replace(contentsUrl, CONTENTS_URL_REPLACE, relativeFilepath, 1)
}

func translateToBranchHistory(branch *github.Branch) *pb.BranchHistory {
	return &pb.BranchHistory{
		Branch: branch.GetName(),
		Hash: branch.Commit.GetSHA(),
		LastCommitTime: &timestamp.Timestamp{Seconds: branch.GetCommit().GetCommit().GetAuthor().GetDate().Unix()},
	}
}

func splitAcctRepo(acctRepo string) (account, repo string) {
	acctRepoList := strings.SplitN(acctRepo, "/", 2)
	account = acctRepoList[0]
	repo = acctRepoList[1]
	return
}

func translateToCommit(commit *github.RepositoryCommit) *pb.Commit {
	return &pb.Commit{
		Hash: commit.GetSHA(),
		Message: commit.GetCommit().GetMessage(),
		Date: &timestamp.Timestamp{Seconds: commit.GetCommit().GetAuthor().GetDate().Unix()},
	}
}