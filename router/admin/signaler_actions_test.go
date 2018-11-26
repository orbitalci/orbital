package admin

import (
	"context"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/go-til/vault"
	"github.com/shankj3/ocelot/build"
	mock_credentials "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/mock_models"
	pbb "github.com/shankj3/ocelot/models/bitbucket/pb"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/status"
)

func TestGuideOcelotServer_WatchRepo(t *testing.T) {
	rc := &vcsRemoteConf{}
	ctx := context.Background()
	handl := &handle{}
	gos := &guideOcelotServer{RemoteConfig: rc, handler: handl}
	acct := &pb.RepoAccount{Repo: "shankj3", Account: "ocelot", Type: pb.SubCredType_GITHUB}
	_, err := gos.WatchRepo(ctx, acct)
	if err == nil {
		t.Error("no hhBaseUrlSet, should error")
	}
	gos.hhBaseUrl = "https://here.me"
	_, err = gos.WatchRepo(ctx, acct)
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
	gos := &guideOcelotServer{Producer: producer, Storage: store}
	poll := &pb.PollRequest{
		Account:  "shankj3",
		Repo:     "ocelot",
		Cron:     "* * * * *",
		Branches: "master,dev",
	}
	// pollExists is false, should insert and write proto msg
	_, err := gos.PollRepo(ctx, poll)
	if err != nil {
		t.Error(err)
	}
	store.failNoCred = true
	_, err = gos.PollRepo(ctx, poll)
	if err == nil {
		t.Fatal("no cred exests")
	}
	if !strings.Contains(err.Error(), "no cred") {
		t.Error("should roll up no cred error")
	}
	store.failNoCred = false
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
	failNoCred	   bool
	pollExists     bool
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

func (s *signalStorage) InsertPoll(account string, repo string, cronString string, branches string, credId int64) error {
	if s.failInsertPoll {
		return errors.New("fail insert poll")
	}
	return nil
}

func (s *signalStorage) RetrieveCred(subCredType pb.SubCredType, identifier, accountName string) (pb.OcyCredder, error) {
	if s.failNoCred {
		return nil, errors.New("no cred")
	}
	return &pb.VCSCreds{Id: 7}, nil
}

type handle struct {
	models.VCSHandler
	failHook   bool
	failDetail bool
}

var detail = pbb.PaginatedRepository_RepositoryValues{
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

func (h *handle) GetRepoLinks(acctRepo string) (*pb.Links, error) {
	if h.failDetail {
		return nil, errors.New("failing detail")
	}
	return &pb.Links{}, nil
}

func (h *handle) SetCallbackURL(string) {}

var ocelot = []byte(`image: jessishank/mavendocker
buildTool: maven
branches:
- master
- banana
env:
  - "MARIANNE=1"
stages:
   - name: inst
     script:
        - mvn clean install
`)

var ocelotInvalid = []byte(`buildTool: maven
branches:
- master
- banana
env:
  - "MARIANNE=1"
stages:
   - name: inst
     script:
        - mvn clean install
`)

func storageSuccessfulBuild(ocelotStorage *storage.MockOcelotStorage, hash, acct, repo, branch string) {
	ocelotStorage.EXPECT().AddSumStart(hash, acct, repo, branch).Return(int64(1), nil).Times(1)
	ocelotStorage.EXPECT().SetQueueTime(int64(1)).Return(nil).Times(1)
	ocelotStorage.EXPECT().AddStageDetail(gomock.Any()).Return(nil).Times(1)
}

func TestGuideOcelotServer_BuildRepoAndHash(t *testing.T) {
	gos, ctl, mockz := setupMockedGuideOcelot(t)
	defer ctl.Finish()
	consl := consul.NewMockConsuletty(ctl)
	vlt := vault.NewMockVaulty(ctl)
	cred := &pb.VCSCreds{ClientSecret: "1", ClientId: "2", Identifier: "BITBUCKET_shankj3", AcctName: "shankj3", SubType: pb.SubCredType_BITBUCKET, TokenURL: "http://token"}
	// test when there is no hash. should look up last commit data for a branch to get the hash to build
	// this should successfully produce a nsq message
	t.Run("no hash build test", func(t *testing.T) {
		// set up expected mock data
		mockz.rc.EXPECT().GetConsul().Return(consl).Times(1)
		mockz.rc.EXPECT().GetVault().Return(vlt).Times(1)
		mockz.rc.EXPECT().GetCred(mockz.store, pb.SubCredType_BITBUCKET, "BITBUCKET_shankj3", "shankj3", false).Return(cred, nil).Times(1)
		mockz.handler.EXPECT().GetBranchLastCommitData("shankj3/ocelot", "banana").Return(&pb.BranchHistory{Hash: "123", Branch: "banana"}, nil)
		mockz.handler.EXPECT().GetFile("ocelot.yml", "shankj3/ocelot", "123").Return(ocelot, nil)
		mockz.handler.EXPECT().GetChangedFiles("shankj3/ocelot", gomock.Any(), gomock.Any()).Return([]string{"ocelot.yml"}, nil)
		mockz.handler.EXPECT().GetCommit("shankj3/ocelot", "123").Return(&pb.Commit{Message:"hi", Hash: "123"}, nil).Times(1)
		consl.EXPECT().GetKeyValue("ci/werker_build_map/123").Return(nil, nil)
		vlt.EXPECT().CreateThrowawayToken().Return("sup", nil)
		mockz.producer.EXPECT().WriteProto(gomock.Any(), "build").Times(1)
		//
		storageSuccessfulBuild(mockz.store, "123", "shankj3", "ocelot", "banana")

		request := &pb.BuildReq{AcctRepo: "shankj3/ocelot", Branch: "banana"}
		streamer := &buildserv{}
		gos.BuildRepoAndHash(request, streamer)
		expectedLines := []string{
			"Searching for VCS creds belonging to shankj3/ocelot...",
			"Successfully found VCS credentials belonging to shankj3/ocelot ✓",
			"Validating VCS Credentials...",
			"Successfully used VCS Credentials to obtain a token ✓",
			"Building branch banana with the latest commit in VCS, which is 123",
			"Retrieving ocelot.yml for shankj3/ocelot...",
			"Successfully retrieved ocelot.yml for shankj3/ocelot ✓",
			"Validating and queuing build data for shankj3/ocelot...",
			"Build started for 123 belonging to shankj3/ocelot ✓",
		}
		if diff := deep.Equal(expectedLines, streamer.lines); diff != nil {
			t.Error(diff)
		}
	})
	//
	//// test scenario where hash and branch are both sent in the build request. an old build summary should be attempted to be retrieved to
	//// validate build info against, and even if it isn't there a build message should be put on the queue
	t.Run("test building with a branch and a hash", func(t *testing.T) {
		request := &pb.BuildReq{AcctRepo: "shankj3/ocelot", Branch: "banana", Hash: "123"}
		mockz.rc.EXPECT().GetConsul().Return(consl).Times(1)
		mockz.rc.EXPECT().GetVault().Return(vlt).Times(1)
		mockz.rc.EXPECT().GetCred(mockz.store, pb.SubCredType_BITBUCKET, "BITBUCKET_shankj3", "shankj3", false).Return(cred, nil).Times(1)
		mockz.store.EXPECT().RetrieveLatestSum("123").Return(&pb.BuildSummary{}, storage.BuildSumNotFound("123"))
		mockz.handler.EXPECT().GetFile("ocelot.yml", "shankj3/ocelot", "123").Return(ocelot, nil)
		mockz.handler.EXPECT().GetChangedFiles("shankj3/ocelot", gomock.Any(), gomock.Any()).Return([]string{"ocelot.yml"}, nil)
		mockz.handler.EXPECT().GetCommit("shankj3/ocelot", "123").Return(&pb.Commit{Message:"hi", Hash: "123"}, nil).Times(1)

		consl.EXPECT().GetKeyValue("ci/werker_build_map/123").Return(nil, nil)
		vlt.EXPECT().CreateThrowawayToken().Return("sup", nil)
		// if this producer expect passes, it means that a message was produced
		mockz.producer.EXPECT().WriteProto(gomock.Any(), "build").Times(1)
		storageSuccessfulBuild(mockz.store, "123", "shankj3", "ocelot", "banana")
		streamer := &buildserv{}
		gos.BuildRepoAndHash(request, streamer)
	})

	//// test scenario where build conf isn't valid
	t.Run("invalid ocelot yml", func(t *testing.T) {
		mockz.rc.EXPECT().GetConsul().Return(consl).Times(1)
		consl.EXPECT().GetKeyValue("ci/werker_build_map/123").Return(nil, nil)
		request := &pb.BuildReq{AcctRepo: "shankj3/ocelot", Branch: "master", Hash: "123"}
		mockz.store.EXPECT().RetrieveLatestSum("123").Return(&pb.BuildSummary{}, storage.BuildSumNotFound("123"))
		mockz.rc.EXPECT().GetCred(mockz.store, pb.SubCredType_BITBUCKET, "BITBUCKET_shankj3", "shankj3", false).Return(cred, nil).Times(1)
		mockz.handler.EXPECT().GetFile("ocelot.yml", "shankj3/ocelot", "123").Return(ocelotInvalid, nil)
		mockz.handler.EXPECT().GetChangedFiles("shankj3/ocelot", gomock.Any(), gomock.Any()).Return([]string{"ocelot.yml"}, nil)
		mockz.handler.EXPECT().GetCommit("shankj3/ocelot", "123").Return(&pb.Commit{Message:"hi", Hash: "123"}, nil).Times(1)
		mockz.store.EXPECT().AddSumStart("123", "shankj3", "ocelot", "master").Return(int64(1), nil).Times(1)
		mockz.store.EXPECT().StoreFailedValidation(int64(1)).Times(1)
		mockz.store.EXPECT().AddStageDetail(gomock.Any()).Return(nil).Times(1)
		streamer := &buildserv{}
		gos.BuildRepoAndHash(request, streamer)
	})
}

func setupMockedGuideOcelot(t *testing.T) (*guideOcelotServer, *gomock.Controller, *mocks) {
	ctl := gomock.NewController(t)
	mockz := &mocks{}
	mockz.store = storage.NewMockOcelotStorage(ctl)
	mockz.handler = mock_models.NewMockVCSHandler(ctl)
	mockz.rc = mock_credentials.NewMockCVRemoteConfig(ctl)
	mockz.producer = nsqpb.NewMockProducer(ctl)
	gos := &guideOcelotServer{
		Storage:      mockz.store,
		RemoteConfig: mockz.rc,
		handler:      mockz.handler,
		Deserializer: deserialize.New(),
		Producer:     mockz.producer,
		OcyValidator: build.GetOcelotValidator(),
	}
	return gos, ctl, mockz
}

type mocks struct {
	producer *nsqpb.MockProducer
	rc       *mock_credentials.MockCVRemoteConfig
	store    *storage.MockOcelotStorage
	handler  *mock_models.MockVCSHandler
}

func TestGuideOcelotServer_BuildRepoAndHash_previouslybuilt(t *testing.T) {
	gos, ctl, mockz := setupMockedGuideOcelot(t)
	defer ctl.Finish()
	consl := consul.NewMockConsuletty(ctl)
	vlt := vault.NewMockVaulty(ctl)

	streamer := &buildserv{}

	cred := &pb.VCSCreds{ClientSecret: "1", ClientId: "2", Identifier: "BITBUCKET_shankj3", AcctName: "shankj3", SubType: pb.SubCredType_BITBUCKET, TokenURL: "http://token"}

	request3 := &pb.BuildReq{AcctRepo: "shankj3/ocelot", Hash: "ks72bas"}

	mockz.rc.EXPECT().GetConsul().Return(consl).Times(1)
	mockz.rc.EXPECT().GetVault().Return(vlt).Times(1)
	mockz.rc.EXPECT().GetCred(mockz.store, pb.SubCredType_BITBUCKET, "BITBUCKET_shankj3", "shankj3", false).Return(cred, nil).Times(1)
	mockz.store.EXPECT().RetrieveLatestSum("ks72bas").Return(&pb.BuildSummary{Hash: "ks72basasdfasdf", Branch: "master", BuildId: 123, Repo: "ocelot", Account: "shankj3"}, nil)
	mockz.handler.EXPECT().GetFile("ocelot.yml", "shankj3/ocelot", "ks72basasdfasdf").Return(ocelot, nil)
	mockz.handler.EXPECT().GetChangedFiles("shankj3/ocelot", gomock.Any(), gomock.Any()).Return([]string{"ocelot.yml"}, nil)
	mockz.handler.EXPECT().GetCommit("shankj3/ocelot", "ks72basasdfasdf").Return(&pb.Commit{Message:"hi", Hash: "ks72basasdfasdf" +
		""}, nil).Times(1)

	consl.EXPECT().GetKeyValue("ci/werker_build_map/ks72basasdfasdf").Return(nil, nil)
	vlt.EXPECT().CreateThrowawayToken().Return("sup", nil)
	mockz.producer.EXPECT().WriteProto(gomock.Any(), "build").Times(1)
	storageSuccessfulBuild(mockz.store, "ks72basasdfasdf", "shankj3", "ocelot", "master")
	gos.BuildRepoAndHash(request3, streamer)
	expected := []string{
		"Searching for VCS creds belonging to shankj3/ocelot...",
		"Successfully found VCS credentials belonging to shankj3/ocelot ✓",
		"Validating VCS Credentials...",
		"Successfully used VCS Credentials to obtain a token ✓",
		"No branch was passed, using `master` from build #123 instead...",
		"Found a previous build starting with hash ks72bas, now building branch master ✓",
		"Retrieving ocelot.yml for shankj3/ocelot...",
		"Successfully retrieved ocelot.yml for shankj3/ocelot ✓",
		"Validating and queuing build data for shankj3/ocelot...",
		"Build started for ks72basasdfasdf belonging to shankj3/ocelot ✓",
	}
	if diff := deep.Equal(expected, streamer.lines); diff != nil {
		t.Error(diff)
	}
	// test scenario where hash not found
	mockz.rc.EXPECT().GetCred(mockz.store, pb.SubCredType_BITBUCKET, "BITBUCKET_shankj3", "shankj3", false).Return(cred, nil).Times(1)
	mockz.store.EXPECT().RetrieveLatestSum("ks72bas").Return(&pb.BuildSummary{}, storage.BuildSumNotFound("ks72bas"))
	request2 := &pb.BuildReq{AcctRepo: "shankj3/ocelot", Hash: "ks72bas"}
	streamer = &buildserv{}
	err := gos.BuildRepoAndHash(request2, streamer)
	expected = []string{
		"Searching for VCS creds belonging to shankj3/ocelot...",
		"Successfully found VCS credentials belonging to shankj3/ocelot ✓",
		"Validating VCS Credentials...",
		"Successfully used VCS Credentials to obtain a token ✓",
		"There are no previous builds starting with hash ks72bas...",
	}
	stats, _ := status.FromError(err)
	expectederr := "Branch is a required field if a previous build starting with the specified hash cannot be found. Please pass the branch flag and try again!"
	if stats.Message() != expectederr {
		t.Error(test.StrFormatErrors("err msg", expectederr, stats.Message()))
	}
	if diff := deep.Equal(expected, streamer.lines); diff != nil {
		t.Error(diff)
	}
}

func TestGuideOcelotServer_BuildRepoAndHash_branchNotfound(t *testing.T) {
	gos, ctl, mockz := setupMockedGuideOcelot(t)
	defer ctl.Finish()
	req := &pb.BuildReq{AcctRepo: "shankj3/ocelot", Branch: "schmorgazbord"}
	cred := &pb.VCSCreds{ClientSecret: "1", ClientId: "2", Identifier: "BITBUCKET_shankj3", AcctName: "shankj3", SubType: pb.SubCredType_BITBUCKET, TokenURL: "http://token"}
	mockz.rc.EXPECT().GetCred(mockz.store, pb.SubCredType_BITBUCKET, "BITBUCKET_shankj3", "shankj3", false).Return(cred, nil).Times(1)
	mockz.handler.EXPECT().GetBranchLastCommitData("shankj3/ocelot", "schmorgazbord").Return(nil, models.Branch404("schmorgazbord", "shankj3/ocelot")).Times(1)
	streamer := &buildserv{}
	errmsg := "Branch schmorgazbord was not found for repository shankj3/ocelot"
	err := gos.BuildRepoAndHash(req, streamer)
	statErr, _ := status.FromError(err)
	if statErr.Message() != errmsg {
		t.Error(test.StrFormatErrors("err msg", errmsg, statErr.Message()))
	}
	expected := []string{
		"Searching for VCS creds belonging to shankj3/ocelot...",
		"Successfully found VCS credentials belonging to shankj3/ocelot ✓",
		"Validating VCS Credentials...",
		"Successfully used VCS Credentials to obtain a token ✓",
	}
	if diff := deep.Equal(expected, streamer.lines); diff != nil {
		t.Error(diff)
	}
}

type buildserv struct {
	pb.GuideOcelot_BuildRepoAndHashServer
	lines []string
}

func (b *buildserv) Send(resp *pb.LineResponse) error {
	b.lines = append(b.lines, resp.OutputLine)
	return nil
}
