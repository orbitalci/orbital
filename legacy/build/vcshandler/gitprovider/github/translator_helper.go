package github

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/go-github/v19/github"

	"github.com/level11consulting/orbitalci/models/pb"
)

func translatePushCommit(commit *github.PushEventCommit) (*pb.Commit) {
	return &pb.Commit{
		Message: commit.GetMessage(),
		Hash: commit.GetID(),
		Date: &timestamp.Timestamp{Seconds: commit.GetTimestamp().Unix()},
		Author: &pb.User{UserName: commit.GetAuthor().GetLogin(), DisplayName: commit.GetAuthor().GetName()},
	}
}

func translatePushCommits(commits []github.PushEventCommit) (pbCommits []*pb.Commit) {
	for _, cmt := range commits {
		pbCommits = append(pbCommits, translatePushCommit(&cmt))
	}
	return
}


// getPreviousHead will check the 'before' field to make see if this push event is a new branch push,
// and build the previousHead field accordingly
// order is as follows:
//   1. if there was a previous head commit, set the hash to that
//   2. if there is no previous head commit but there are commits in the push, set the
//		previous head to the last commit
//   3. previous head is null
func getPreviousHead(push *github.PushEvent) (previousHead *pb.Commit) {
	// if there was a previous head commit, set the hash to that
	if push.GetBefore() != newBranchBefore {
		previousHead = &pb.Commit{Hash: push.GetBefore()}
	} else if len(push.Commits) != 0 {
		// if there is no previous head commit but there are commits in the push, set the
		// previous head to the last commit
		previousHead = translatePushCommit(&push.Commits[0])
	}
	return
}

//getPrUrlsFromPR will pull out the urls provided in the pr event and use it to instantiate a PrUrls object
// *DECLINE, APPROVE, AND MERGE WILL NOT BE RENDERED* because that isn't present in github PR event objects
func getPrUrlsFromPR(pr *github.PullRequestEvent) *pb.PrUrls {
	return &pb.PrUrls{
		Commits: pr.PullRequest.GetCommitsURL(),
		Comments: pr.PullRequest.GetCommentsURL(),
		Statuses: pr.PullRequest.GetStatusesURL(),
		// writing these explicitly so it is obcious that you will not get the decline/approve/merge urls
		Decline: "",
		Approve: "",
		Merge: "",
	}
	return nil
}

func translateHeadData(headData *github.PullRequestBranch) *pb.HeadData {
	return &pb.HeadData{
		Branch: headData.GetRef(),
		Hash: headData.GetSHA(),
		Repo: &pb.Repo{
			Name: headData.GetRepo().GetName(),
			AcctRepo: headData.GetRepo().GetFullName(),
			RepoLink: headData.GetRepo().GetURL(),
		},
	}
}
