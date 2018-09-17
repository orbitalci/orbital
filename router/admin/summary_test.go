package admin

import (
	"context"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGuideOcelotServer_LastFewSummaries(t *testing.T) {
	store := &summstorage{}
	gos := &guideOcelotServer{Storage: store}
	repoAct := &pb.RepoAccount{Repo: "shankj3", Account: "ocelot", Limit: 1}
	ctx := context.Background()
	sums, err := gos.LastFewSummaries(ctx, repoAct)
	if err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(sums.Sums[0], pbsummary); diff != nil {
		t.Error(diff)
	}
	repoAct.Repo = ""
	_, err = gos.LastFewSummaries(ctx, repoAct)
	if err == nil {
		t.Error("repo is nil, should return error")
	}
	repoAct.Repo = "shankj3"
	store.returnErr = true
	_, err = gos.LastFewSummaries(ctx, repoAct)
	if err == nil {
		t.Error("store returned an error, should not be nil")
	}
	store.returnEmpty = true
	store.returnErr = false
	_, err = gos.LastFewSummaries(ctx, repoAct)
	if err == nil {
		t.Error("store returned an empty array, should not be nil")
	}
	sErr, ok := status.FromError(err)
	if !ok {
		t.Error("hsould return admin grpc error")
	}
	if sErr.Code() != codes.NotFound {
		t.Error("should return not found as it returned an emptryarray")
	}

}

type summstorage struct {
	returnErr   bool
	notFound    bool
	returnEmpty bool
	storage.OcelotStorage
}

var summary = &pb.BuildSummary{
	Hash:          "hash",
	Failed:        true,
	Account:       "shankj3",
	Repo:          "ocelot",
	Branch:        "master",
	BuildId:       12,
	BuildDuration: 12.1234,
	QueueTime:     &timestamp.Timestamp{Seconds: time.Unix(0, 0).Unix()},
	BuildTime:     &timestamp.Timestamp{Seconds: time.Unix(0, 0).Unix()},
}

var pbsummary = &pb.BuildSummary{
	Hash:          summary.Hash,
	Failed:        summary.Failed,
	Account:       summary.Account,
	Repo:          summary.Repo,
	Branch:        summary.Branch,
	BuildId:       summary.BuildId,
	BuildDuration: summary.BuildDuration,
	QueueTime:     &timestamp.Timestamp{Seconds: 0},
	BuildTime:     &timestamp.Timestamp{Seconds: 0},
}

func (s *summstorage) RetrieveLastFewSums(repo string, account string, limit int32) ([]*pb.BuildSummary, error) {
	if s.returnErr {
		return nil, errors.New("returing an error")
	}
	if s.notFound {
		return nil, storage.BuildSumNotFound(repo)
	}
	if s.returnEmpty {
		return []*pb.BuildSummary{}, nil
	}
	var sums []*pb.BuildSummary
	for i := 0; i < int(limit); i++ {
		sums = append(sums, summary)
	}
	return sums, nil
}
