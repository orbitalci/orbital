package bitbucket

import (
	"errors"
	"io"

	"github.com/golang/protobuf/jsonpb"
	ocelog "github.com/shankj3/go-til/log"
	pbb "github.com/shankj3/ocelot/models/bitbucket/pb"
	"github.com/shankj3/ocelot/models/pb"
)

func GetTranslator() *BBTranslate {
	return &BBTranslate{Unmarshaler: jsonpb.Unmarshaler{AllowUnknownFields: true}}
}

type BBTranslate struct {
	Unmarshaler jsonpb.Unmarshaler
}

func (bb *BBTranslate) TranslatePR(reader io.Reader) (*pb.PullRequest, error) {
	pr := &pbb.PullRequest{}
	err := bb.Unmarshaler.Unmarshal(reader, pr)
	if err != nil {
		return nil, err
	}
	prN := &pb.PullRequest{
		Id:          pr.Pullrequest.Id,
		Description: pr.Pullrequest.Description,
		Urls: &pb.PrUrls{
			Commits:  pr.Pullrequest.Links.Commits.Href,
			Comments: pr.Pullrequest.Links.Comments.Href,
			Statuses: pr.Pullrequest.Links.Statuses.Href,
			Approve:  pr.Pullrequest.Links.Approve.Href,
			Decline:  pr.Pullrequest.Links.Decline.Href,
			Merge:    pr.Pullrequest.Links.Merge.Href,
		},
		Title: pr.Pullrequest.Title,
		Source: &pb.HeadData{
			Repo: &pb.Repo{
				Name:     pr.Pullrequest.Source.Repository.FullName,
				AcctRepo: pr.Pullrequest.Source.Repository.FullName,
				RepoLink: pr.Pullrequest.Source.Repository.Links.Html.Href,
			},
			Branch: pr.Pullrequest.Source.Branch.GetName(),
			Hash:   pr.Pullrequest.Source.Commit.Hash,
		},
		Destination: &pb.HeadData{
			Repo: &pb.Repo{
				Name:     pr.Pullrequest.Destination.Repository.FullName,
				AcctRepo: pr.Pullrequest.Destination.Repository.FullName,
				RepoLink: pr.Pullrequest.Destination.Repository.Links.Html.Href,
			},
			Branch: pr.Pullrequest.Destination.Branch.GetName(),
			Hash:   pr.Pullrequest.Destination.Commit.Hash,
		},
	}
	return prN, nil
}

func (bb *BBTranslate) TranslatePush(reader io.Reader) (*pb.Push, error) {
	push := &pbb.RepoPush{}
	err := bb.Unmarshaler.Unmarshal(reader, push)
	if err != nil {
		return nil, err
	}

	if len(push.Push.Changes) < 1 {
		ocelog.Log().Error("no commits found in push")
		// todo: evaluate if this is actually what should happen, what about pushing tags?
		return nil, errors.New("no commits found in push")
	}
	if len(push.Push.Changes) > 1 {
		ocelog.Log().Errorf("length of push changes is > 1, changes are %#v", push.Push.Changes)
		return nil, errors.New("too many changesets")
	}
	var commits []*pb.Commit
	//todo: when is there ever more than one changeset in the changes array?
	changeset := push.Push.Changes[0]
	var last int
	for ind, commit := range changeset.Commits {
		commits = append(commits, &pb.Commit{Hash: commit.Hash, Date: commit.Date, Message: commit.Message, Author: &pb.User{UserName: commit.Author.Username}})
		last = ind
	}
	if changeset.New.Type != "branch" {
		ocelog.Log().Errorf("changeset type is not branch, it is %s. idk what to do!!!", changeset.New.Type)
		ocelog.Log().Error(push)
		return nil, errors.New("unexpected push type")
	}
	if changeset.New.Target.Hash != commits[0].Hash {
		ocelog.Log().WithField("changeset.New.Target.Hash", changeset.New.Target.Hash).WithField("commits[0].Hash", commits[0].Hash).Error("WHAT THE HELL? new & first commit SHOULD BE THE SAME!")
	}
	if changeset.Old.Target.Hash != commits[last].Hash {
		ocelog.Log().Infof("setting last commit to be %s, as it was %s", changeset.Old.Target.Hash, commits[last].Hash)
		commits = append(commits, &pb.Commit{Hash: changeset.Old.Target.Hash, Message: changeset.Old.Target.Message, Author: &pb.User{UserName: changeset.Old.Target.Author.Username}, Date: changeset.Old.Target.Date})
	}
	translPush := &pb.Push{
		Repo:               &pb.Repo{Name: push.Repository.FullName, AcctRepo: push.Repository.FullName, RepoLink: push.Repository.Links.Html.Href},
		User:               &pb.User{UserName: push.Actor.Username},
		Commits:            commits,
		Branch:             changeset.New.Name,
		HeadCommit:         &pb.Commit{Hash: changeset.New.Target.Hash, Message: changeset.New.Target.Message, Date: changeset.New.Target.Date, Author: &pb.User{UserName: changeset.New.Target.Author.Username}},
		PreviousHeadCommit: &pb.Commit{Hash: changeset.Old.Target.Hash, Message: changeset.Old.Target.Message, Date: changeset.Old.Target.Date, Author: &pb.User{UserName: changeset.Old.Target.Author.Username}},
}
	return translPush, nil
}
