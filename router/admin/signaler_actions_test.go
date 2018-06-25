package admin

import (
	"context"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/models"
	pbb "github.com/shankj3/ocelot/models/bitbucket/pb"
	"github.com/shankj3/ocelot/models/pb"
)

func TestGuideOcelotServer_WatchRepo(t *testing.T) {
	rc := &vcsRemoteConf{}
	ctx := context.Background()
	handl := &handle{}
	gos := &guideOcelotServer{RemoteConfig:rc, handler:handl}
	acct := &pb.RepoAccount{Repo: "shankj3", Account: "ocelot"}
	_, err := gos.WatchRepo(ctx, acct)
	if err != nil {
		t.Error(err)
	}
	handl.failHook = true
	_, err = gos.WatchRepo(ctx, acct)
	if err == nil {
		t.Error("failed to create webhoook, should fail.")
	}
	if !strings.Contains(err.Error(), "failing webhoook") {
		t.Error("should sho webhook error, instead showing ", err.Error())
	}
	handl.failHook = false
	handl.failDetail = true
	_, err = gos.WatchRepo(ctx, acct)
	if err == nil {
		t.Error("failed to get acctrepo detail, should fail.")
	}
	handl.failDetail = false
	rc.returnErr = true
	_, err = gos.WatchRepo(ctx, acct)
	if err == nil {
		t.Error("could not get vcs creds, shoudl fail")
	}
	rc.returnErr = false

	acct.Account = ""
	_, err = gos.WatchRepo(ctx, acct)
	if err == nil {
		t.Error("account is empty, should return error")
	}
	acct.Account = "shankj3"
	acct.Repo = ""
	_, err = gos.WatchRepo(ctx, acct)
	if err == nil {
		t.Error("repo is empty, should return error")
	}
	gos.handler = nil
	acct.Repo = "ocelot"
	_, err = gos.WatchRepo(ctx, acct)
	if err == nil {
		t.Error("repo is empty, should return error")
	}
	if !strings.Contains(err.Error(), "Unable to retrieve the bitbucket client config f") {
		t.Error("shuld return handler error")
	}
}

func TestGuideOcelotServer_PollRepo(t *testing.T) {
	brs := &buildruntimestorage{}
	store := &signalStorage{buildruntimestorage: brs}
	producer := &fakeProducer{}
	ctx := context.Background()
	gos := &guideOcelotServer{Producer:producer, Storage:store}
	poll := &pb.PollRequest{
		Account: "shankj3",
		Repo:"ocelot",
		Cron: "* * * * *",
		Branches: "master,dev",
	}
	// pollExists is false, should insert and write proto msg
	_, err := gos.PollRepo(ctx, poll)
	if err != nil {
		t.Error(err)
	}
	store.failInsertPoll = true
	_, err = gos.PollRepo(ctx, poll)
	if err == nil {
		t.Error("storage is failing on inserting poll, should fail")
	}
	if !strings.Contains(err.Error(), "unable to insert poll into storage") {
		t.Error("should return could not insert poll error")
	}
	//reset
	store.failInsertPoll = false
	// poll does exist, should update happily
	store.pollExists = true
	_, err = gos.PollRepo(ctx, poll)
	if err != nil {
		t.Error(err)
	}
	store.failUpdatePoll = true
	_, err = gos.PollRepo(ctx, poll)
	if err == nil {
		t.Error("storage is failing on updating poll, should fail")
	}
	if !strings.Contains(err.Error(), "unable to update poll in storage") {
		t.Error("should return could not update poll error, returned: " + err.Error())
	}
	//reset
	store.failUpdatePoll = false

	store.failPollExists = true
	_, err = gos.PollRepo(ctx, poll)
	if err == nil {
		t.Error("storage rcould not check if poll exists, should fail")
	}
	if !strings.Contains(err.Error(), "unable to retrieve poll table from storage. ") {
		t.Error("should return error about not retrieving poll table, returned: " + err.Error())
	}
	//reset
	store.failPollExists = false
	empty := &pb.PollRequest{}
	_, err = gos.PollRepo(ctx, empty)
	if err == nil {
		t.Error("no reqeust params sent, should return error")
	}
	if !strings.Contains(err.Error(), "account, repo, cron, and branches are required fields") {
		t.Error("should return validation error, returned: " + err.Error())
	}

	producer.returnErr = true
	_, err = gos.PollRepo(ctx, poll)
	if err == nil {
		t.Error("producer returned error, this should fail")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Error("should return error from produecer, returend: " + err.Error())
	}

}

type signalStorage struct {
	*buildruntimestorage
	failUpdatePoll bool
	failInsertPoll bool
	failPollExists bool
	pollExists bool
}

func (s *signalStorage) PollExists(account string, repo string) (bool, error) {
	if s.failPollExists {
		return false, errors.New("failing exists")
	}
	return s.pollExists, nil
}


func (s *signalStorage) UpdatePoll(account string, repo string, cronString string, branches string) error {
	if s.failUpdatePoll {
		return errors.New("fail update poll")
	}
	return nil
}

func (s *signalStorage) InsertPoll(account string, repo string, cronString string, branches string) error {
	if s.failInsertPoll {
		return errors.New("fail insert poll")
	}
	return nil
}



type handle struct {
	models.VCSHandler
	failHook bool
	failDetail bool
}

var detail = pbb.PaginatedRepository_RepositoryValues {
	Type: "repo",
	Links: &pbb.PaginatedRepository_RepositoryValues_RepositoryLinks{
		Hooks: &pbb.LinkUrl{
			Href: "http://webhook.forever/yo",
		},
	},
}

func (h *handle) GetRepoDetail(acctRepo string) (pbb.PaginatedRepository_RepositoryValues, error) {
	if h.failDetail {
		return pbb.PaginatedRepository_RepositoryValues{}, errors.New("failing detail")
	}
	return detail, nil
}

func (h *handle) CreateWebhook(webhookURL string) error {
	if h.failHook {
		return errors.New("failing webhoook")
	}
	return nil
}