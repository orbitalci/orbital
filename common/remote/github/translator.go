package github

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/google/go-github/v19/github"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/log"
	"github.com/level11consulting/ocelot/models"

	"github.com/level11consulting/ocelot/models/pb"
)

// newBranchBefore is the "before" field in github if a push event is a new branch
const newBranchBefore = "0000000000000000000000000000000000000000"
const branchesRefPrefix = "refs/heads/"

func GetTranslator() models.Translator {
	return &translator{}
}

type translator struct {}


func (t *translator) TranslatePush(reader io.Reader) (*pb.Push, error) {
	bytez, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read push event into bytes")
	}
	hookPush, err := github.ParseWebHook("push", bytez)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse push event into github struct")
	}
	push, ok := hookPush.(*github.PushEvent)
	if !ok {
		return nil, errors.New("unable to cast as github.PushEvent")
	}
	if !strings.Contains(push.GetRef(), branchesRefPrefix) {
		log.Log().Errorf("changeset type is not branch, it is %s. idk what to do!!!", push.GetRef())
		return nil, errors.New("unexpected push type")
	}
	branch := strings.Replace(push.GetRef(), branchesRefPrefix, "", 1)
	pbPush := &pb.Push{
		Repo: &pb.Repo{
			Name: push.GetRepo().GetName(),
			AcctRepo: push.GetRepo().GetFullName(),
			RepoLink: push.GetRepo().GetURL(),
		},
		User: &pb.User{UserName: push.GetRepo().GetOwner().GetLogin()},
		HeadCommit: translatePushCommit(push.GetHeadCommit()),
		PreviousHeadCommit: getPreviousHead(push),
		Commits: translatePushCommits(push.Commits),
		Branch: branch,
	}
	return pbPush, nil
}

func (t *translator) TranslatePR(reader io.Reader) (*pb.PullRequest, error) {
	bytez, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read push event into bytes")
	}
	hookPR, err := github.ParseWebHook("pull_request", bytez)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse pr event into github struct")
	}
	pr := (hookPR).(*github.PullRequestEvent)
	// if this webhook was fired because a PR was merged, it shouldn't trigger a build because the merge
	// will generate a push event with the new head hash, and _that_ should be built
	if pr.GetAction() == "closed" {
		return nil, models.DontBuildEvent(pb.SubCredType_GITHUB, pr.GetAction())
	}

	pullReq := &pb.PullRequest{
		Urls: getPrUrlsFromPR(pr),
		Description: pr.PullRequest.GetBody(),
		Title: pr.PullRequest.GetTitle(),
		Source: translateHeadData(pr.PullRequest.GetHead()),
		Destination: translateHeadData(pr.PullRequest.GetBase()),
		Id: int64(pr.GetNumber()),
	}
	return pullReq, nil
}
