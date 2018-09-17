package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stor struct {
	returnErr bool
	notFound  bool
	storage.OcelotStorage
}

func (s *stor) GetTrackedRepos() (*pb.AcctRepos, error) {
	if s.returnErr {
		return nil, errors.New("failing tracked repos")
	}
	if s.notFound {
		return nil, storage.BuildSumNotFound("all")
	}
	return &pb.AcctRepos{
		AcctRepos: []*pb.AcctRepo{
			{Account: "shankj3", Repo: "123"},
			{Account: "shankj3", Repo: "12323"},
			{Account: "shankj3", Repo: "repo"},
		}}, nil
}

func TestGuideOcelotServer_GetTrackedRepos(t *testing.T) {
	stora := &stor{}
	gos := &guideOcelotServer{Storage: stora}
	ctx := context.Background()
	repos, err := gos.GetTrackedRepos(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if repos.AcctRepos[0].Repo != "123" {
		t.Error("bad data")
	}
	stora.notFound = true
	_, err = gos.GetTrackedRepos(ctx, nil)
	if err == nil {
		t.Error("should return not found error")
	}
	statErr, ok := status.FromError(err)
	if !ok {
		t.Error("must return admin grpc error")
	}
	if statErr.Code() != codes.NotFound {
		t.Error("stiorage returned a not found erorr, grpc code should be set to NOT FOUND")
	}
	stora.notFound = false
	stora.returnErr = true
	_, err = gos.GetTrackedRepos(ctx, nil)
	if err == nil {
		t.Error("should return error")
	}
	statErr, ok = status.FromError(err)
	if !ok {
		t.Error("must return admin grpc error")
	}
	if statErr.Message() != "an error occurred getting account/repos from db" {
		t.Error(statErr.Message(), " is not a correct error message.")
	}
}
