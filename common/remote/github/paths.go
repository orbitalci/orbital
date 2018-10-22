package github

import (
	"fmt"
	"strings"
)

var (
	// all repositories of authenticated user
	ALLREPOS = "/user/repos"
	REPOS = "/repos"
	BRANCHES = REPOS + "/%s/branches%s"
	ISSUE_COMMENTS = REPOS + "/%s/issues/%s/comments"
	FILE = REPOS + "/%s/contents/%s"
	COMMIT = "/repos/%s/commits%s"
	// compare url GET /repos/:owner/:repo/compare/:base...:head
	COMPARE = "/repos/%s/compare/%s...%s"

)

// url replacements for github, see below for urls returned and what to replace

var (
	//"contents_url": "https://api.github.com/repos/shankj3/legis_data/contents/{+path}",
	CONTENTS_URL_REPLACE = "{+path}"
	//"commits_url": "https://api.github.com/repos/shankj3/legis_data/commits{/sha}",
	COMMITS_URL_REPLACE = "{/sha}"
	//"git_commits_url": "https://api.github.com/repos/shankj3/legis_data/git/commits{/sha}",
	GIT_COMMITS_URL_REPLACE = COMMITS_URL_REPLACE
	//"compare_url": "https://api.github.com/repos/shankj3/legis_data/compare/{base}...{head}",
	COMPARE_URL_BASE_REPLACE = "{base}"
	//"compare_url": "https://api.github.com/repos/shankj3/legis_data/compare/{base}...{head}",
	COMPARE_URL_HEAD_REPLACE = "{head}"
	//"pulls_url": "https://api.github.com/repos/shankj3/legis_data/pulls{/number}",
	PULLS_URL_REPLACE = "{/number}"
	//"hooks_url": "https://api.github.com/repos/shankj3/lego/hooks",
)

func getUrlForFileFromContentsUrl(contentsUrl string, relativeFilepath string) string {
	return strings.Replace(contentsUrl, CONTENTS_URL_REPLACE, relativeFilepath, 1)
}

func getUrlForHooksFromHooksUrl(hooksUrl, hookId string) string {
	if hookId != "" {
		hookId = "/" + hookId
	}
	return hooksUrl + hookId
}

func getUrlForCommitsFromCommitsUrl(commitsUrl, hash string) string {
	if hash != "" {
		hash = "/" + hash
	}
	return strings.Replace(commitsUrl, COMMITS_URL_REPLACE, hash, 1)
}

func buildFilePath(accountrepo, filepath string) string {
	return fmt.Sprintf(FILE, accountrepo, filepath)
}

func buildRepoPath(accountRepo string) string {
	return REPOS + "/" + accountRepo
}

func buildBranchesPath(accountRepo, branch string) string {
	if branch != "" {
		branch = "/" + branch
	}
	return fmt.Sprintf(BRANCHES, accountRepo, branch)
}

func buildCommitsPath(acctRepo, singleCommit string) string {
	if singleCommit != "" {
		singleCommit = "/" + singleCommit
	}
	return fmt.Sprintf(COMMIT, acctRepo, singleCommit)
}

func buildPrCommentsPath(acctRepo, prId string) string {
	return fmt.Sprintf(ISSUE_COMMENTS, acctRepo, prId)
}

func buildComparePath(acctRepo, previousHead, newHead string) string {
	return fmt.Sprintf(COMPARE, acctRepo, previousHead, newHead)
}