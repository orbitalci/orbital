package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


func TestGuideOcelotServer_DeletePollRepo(t *testing.T) {
	producer := &fakeProducer{}
	store := &fakeStorage{}
	gos := &guideOcelotServer{Producer: producer, Storage:store}
	ctx := context.Background()
	_, err := gos.DeletePollRepo(ctx, &pb.PollRequest{Account:"1", Repo:"2"})
	if err != nil {
		t.Error("should not fail, erro is: " + err.Error())
	}
	_, err = gos.DeletePollRepo(ctx, &pb.PollRequest{Account:"1"})
	if err == nil {
		t.Error("shouldn't take only account, repo is also required")
	}
	producer.returnErr = true
	_, err = gos.DeletePollRepo(ctx, &pb.PollRequest{Account:"1", Repo:"2"})
	if err == nil {
		t.Error("shuld return error as producer returned an error")
	}
	store.returnErr = true
	producer.returnErr = false
	_, err = gos.DeletePollRepo(ctx, &pb.PollRequest{Account:"1", Repo:"2"})
	if err != nil {
		t.Error("a bad storage shouldn't return an error, as the nsq should still be written to")
	}
}

func TestGuideOcelotServer_ListPolledRepos(t *testing.T) {
	store := &fakeStorage{}
	gos := &guideOcelotServer{Storage:store}
	ctx := context.Background()
	polls, err := gos.ListPolledRepos(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if len(polls.Polls) != 3 {
		t.Error("should have all three poll repos returned")
	}
	store.notFound = true
	_, err = gos.ListPolledRepos(ctx, nil)
	if err == nil {
		t.Error("storage returned not found this should be bubbled up")
	}
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Error("must be grpc error")
	}
	if statusErr.Code() != codes.NotFound {
		t.Error("should return not found code if storage returend not found")
	}
	store.notFound = false
	store.returnErr = true
	_, err = gos.ListPolledRepos(ctx, nil)
	if err == nil {
		t.Error("storage returned unknown error this should be bubbled up")
	}
}

type fakeProducer struct {
	returnErr bool
}

func (f *fakeProducer) WriteProto(message proto.Message, topicName string) error {
	if f.returnErr {
		return errors.New("bad")
	}
	return nil
}

type fakeStorage struct {
	returnErr bool
	notFound bool
	storage.OcelotStorage
}

func (f *fakeStorage) DeletePoll(account string, repo string) error {
	if f.returnErr {
		return errors.New("no delete poll for u")
	}
	return nil
}

var polls = []*models.PollRequest{
	{Account:"shankj3", Repo:"repo2", Cron: "* * * * *", Branches:"branch1,branch2"},
	{Account:"shankj3", Repo:"repo2", Cron: "* * * * *", Branches:"branch1,branch9"},
	{Account:"accountBoiiii", Repo:"repo1", Cron: "* * * * *", Branches:"master,dev"},
}

func (f *fakeStorage) GetAllPolls() ([]*models.PollRequest, error){
	if f.returnErr {
		return nil, errors.New("no get all polls for u")
	}
	if f.notFound {
		return nil, &storage.ErrNotFound{}
	}
	return polls, nil
}
