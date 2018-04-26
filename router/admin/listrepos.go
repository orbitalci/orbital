package admin

import (
	"context"

	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *guideOcelotServer) GetTrackedRepos(ctx context.Context, empty *empty.Empty) (*pb.AcctRepos, error) {
	repos, err := g.Storage.GetTrackedRepos()
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, "could not find any account/repos in the database")
		}
		return nil, status.Error(codes.FailedPrecondition, "an error occurred getting account/repos from db")
	}
	return repos, nil
}