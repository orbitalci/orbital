package admin

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shankj3/go-til/log"

	"github.com/shankj3/ocelot/models/pb"
)

func (g *guideOcelotServer) LastFewSummaries(ctx context.Context, repoAct *pb.RepoAccount) (*pb.Summaries, error) {
	if repoAct.Repo == "" || repoAct.Account == "" || repoAct.Limit == 0 {
		return nil, status.Error(codes.InvalidArgument, "repo, account, and limit are required fields")
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
		summaries.Sums = append(summaries.Sums, model)
	}
	//fmt.Println(summaries)
	return summaries, nil

}
