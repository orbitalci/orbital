package admin

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shankj3/go-til/log"

	"github.com/shankj3/ocelot/models/pb"
)

func (g *guideOcelotServer) LastFewSummaries(ctx context.Context, repoAct *pb.RepoAccount) (*pb.Summaries, error) {
	if repoAct.Repo == "" || repoAct.Account == "" || repoAct.Limit == 0 {
		return nil, status.Error(codes.InvalidArgument, "repo, account, and limit are required fields")
	}
	log.Log().Debug("getting last few summaries")
	if repoAct.Limit == 0 || repoAct.Account == "" || repoAct.Repo == "" {
		return nil, status.Error(codes.InvalidArgument, "limit, account, and repo are all required")
	}
	var summaries = &pb.Summaries{}
	modelz, err := g.Storage.RetrieveLastFewSums(repoAct.Repo, repoAct.Account, repoAct.Limit)
	if err != nil {
		return nil, handleStorageError(err)
	}
	log.Log().Debug("successfully retrieved last few summaries")
	if len(modelz) == 0 {
		return nil, status.Error(codes.NotFound, "no entries found")
	}
	for _, model := range modelz {
		summary := &pb.BuildSummary{
			Hash:          model.Hash,
			Failed:        model.Failed,
			BuildTime:     &timestamp.Timestamp{Seconds: model.BuildTime.UTC().Unix()},
			QueueTime:     &timestamp.Timestamp{Seconds: model.QueueTime.UTC().Unix()},
			Account:       model.Account,
			BuildDuration: model.BuildDuration,
			Repo:          model.Repo,
			Branch:        model.Branch,
			BuildId:       model.BuildId,
		}
		summaries.Sums = append(summaries.Sums, summary)
	}
	fmt.Println(summaries)
	return summaries, nil

}
