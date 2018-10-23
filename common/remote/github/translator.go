package github

//type Translator interface {
//	//TranslatePush should take a reader body, unmarshal it to vcs-specific model, then translate it to the global Push object
//	TranslatePush(reader io.Reader) (*pb.Push, error)
//
//	//TranslatePush should take a reader body, unmarshal it to vcs-specific model, then translate it to the global PullRequest object
//	TranslatePR(reader io.Reader) (*pb.PullRequest, error)
//}
import (
	"io"
	"io/ioutil"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"

	"github.com/shankj3/ocelot/models/pb"
)

type translator struct {

}

func translatePushCommit(commit *github.PushEventCommit) (*pb.Commit) {
	return &pb.Commit{
		Message: commit.GetMessage(),
		Hash: commit.GetSHA(),
		Date: &timestamp.Timestamp{Seconds: commit.GetTimestamp().Unix()},
		Author: &pb.User{UserName: commit.GetAuthor().GetName()},
	}
}

func (t *translator) TranslatePush(reader io.Reader) (*pb.Push, error) {
	bytez, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read push event into bytes")
	}
	hookPush, err := github.ParseWebHook("push", bytez)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse push event into github struct")
	}
	push, err := hookPush.(*github.PushEvent)
	if err != nil {
		return nil, errors.Wrap(err, "unable to cast as github.PushEvent")
	}
	var commits []*pb.Commit
	for _, cmt := range push.Commits {
		commits = append(commits, translatePushCommit(&cmt))
	}
	pbPush := &pb.Push{
		Repo: &pb.Repo{
			Name: push.GetRepo().GetName(),
			AcctRepo: push.GetRepo().GetFullName(),
			RepoLink: push.GetRepo().GetURL(),
		},
		User: &pb.User{UserName: push.GetRepo().GetOwner().GetName()},
		HeadCommit: translatePushCommit(push.GetHeadCommit()),
		PreviousHeadCommit: &pb.Commit{Hash: push.GetBefore()},
		Commits: commits,
		Branch: push.
	}

}
