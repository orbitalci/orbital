package commiteventhandler

//import (
//	"io/ioutil"
//	"testing"
//)
//
//func TestRepoPush(t *testing.T) {
//	//repoPushData := ioutil.ReadAll("test-fixtures/")
//}

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/timestamp"
	dto "github.com/prometheus/client_model/go"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/build_signaler"
	"github.com/level11consulting/ocelot/common/credentials"
	"github.com/level11consulting/ocelot/models/mock_models"

	//"github.com/level11consulting/ocelot/models/mock_models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
)

func createMockedHHC(t *testing.T) (*HookHandlerContext, *credentials.MockCVRemoteConfig, *nsqpb.MockProducer, *storage.MockOcelotStorage, *mock_models.MockVCSHandler, *build_signaler.MockCommitPushWerkerTeller, *build_signaler.MockPRWerkerTeller){
	ctl := gomock.NewController(t)
	handler := mock_models.NewMockVCSHandler(ctl)
	rc := credentials.NewMockCVRemoteConfig(ctl)
	produce := nsqpb.NewMockProducer(ctl)
	store := storage.NewMockOcelotStorage(ctl)
	teller := build_signaler.NewMockCommitPushWerkerTeller(ctl)
	prTeller := build_signaler.NewMockPRWerkerTeller(ctl)
	hhc := &HookHandlerContext{
		Signaler: &build_signaler.Signaler{
			RC: 	      rc,
			Producer:     produce,
			Store: 	      store,
			Deserializer: deserialize.New(),
			OcyValidator: build.GetOcelotValidator(),
		},
		testingHandler: handler,
		pTeller: teller,
		prTeller: prTeller,
	}
	return hhc, rc, produce, store, handler, teller, prTeller
}

var jessiUser = &pb.User{
	UserName:    "jessi_shank",
	DisplayName: "Jessi Shank",
}

var twoPushCommits = []*pb.Commit{
	{
		Hash: "4f59e57d8878daabcbc6af6bd374929b9fe03568",
		Message: "adding test data\n",
		Author: jessiUser,
		Date: &timestamp.Timestamp{Seconds:1537998853},
	},
	{
		Hash: "48bd5377046bdd5bcce847c50b3bce90e22bd66d",
		Message: "another one!\n",
		Author: jessiUser,
		Date: &timestamp.Timestamp{Seconds:1537998814},
	},
	{
		Hash: "906a8969084b9ab0d8ce1056d3e046629eb2d6f9",
		Message: "adding one more change after submitting pr!!! woooahh nelly!\n",
		Author: jessiUser,
		Date: &timestamp.Timestamp{Seconds:1537998072},
	},
}

//var jessiUser = &pb.User{DisplayName: "Jessi Shank", UserName: "jessi_shank"}
var jessiOwner = &pb.User{DisplayName: "Jessi Shank", UserName: "jessishank"}
var mytestOcy = &pb.Repo{
	Name: "jessishank/mytestocy",
	RepoLink: "https://bitbucket.org/jessishank/mytestocy",
	AcctRepo: "jessishank/mytestocy",
}

func TestRepoPush_happy(t *testing.T) {
	twoCommitPushFp := filepath.Join("test-fixtures", "two_commit_push.json")

	twoCommitPush, err := os.Open(twoCommitPushFp)
	if err != nil {
		t.Error(err)
	}
	req := httptest.NewRequest("POST", "/bitbucket", twoCommitPush)
	resp := httptest.NewRecorder()

	hhc, mockRC, _, mockStore, handler, teller, _ := createMockedHHC(t)
	ident, _ := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, "jessishank")
	mockRC.EXPECT().GetCred(mockStore, pb.SubCredType_BITBUCKET, ident, "jessishank", false).Return(&pb.VCSCreds{AcctName:"jessishank", ClientSecret:"xxx", Identifier: ident}, nil).Times(1)
	push := &pb.Push{
		Repo: mytestOcy,
		User: jessiOwner,
		Branch: "here_we_go",
		HeadCommit: &pb.Commit{
			Hash: "4f59e57d8878daabcbc6af6bd374929b9fe03568",
			Message: "adding test data\n",
			Date: &timestamp.Timestamp{Seconds:1537998853},
			Author: jessiUser,
		},
		PreviousHeadCommit: &pb.Commit{
			Hash: "906a8969084b9ab0d8ce1056d3e046629eb2d6f9",
			Message: "adding one more change after submitting pr!!! woooahh nelly!\n",
			Date: &timestamp.Timestamp{Seconds:1537998072},
			Author: jessiUser,
		},
		Commits: []*pb.Commit{
			{
				Hash:    "4f59e57d8878daabcbc6af6bd374929b9fe03568",
				Message: "adding test data\n",
				Date: &timestamp.Timestamp{Seconds:1537998853},
				Author:  jessiUser,
			},
			{
				Hash:    "48bd5377046bdd5bcce847c50b3bce90e22bd66d",
				Message: "another one!\n",
				Author:  jessiUser,
				Date: &timestamp.Timestamp{Seconds:1537998814},
			},
		},

	}

	teller.EXPECT().TellWerker(push, hhc.Signaler, handler, "token", false, pb.SignaledBy_PUSH).Return(nil).Times(1)
	hhc.RepoPush(resp, req, pb.SubCredType_BITBUCKET)
}

func TestRepoPush_credFailed(t *testing.T) {
	twoCommitPushFp := filepath.Join("test-fixtures", "two_commit_push.json")

	twoCommitPush, err := os.Open(twoCommitPushFp)
	if err != nil {
		t.Error(err)
	}
	req := httptest.NewRequest("POST", "/bitbucket", twoCommitPush)
	resp := httptest.NewRecorder()

	hhc, mockRC, _, mockStore, _, _, _ := createMockedHHC(t)
	ident, _ := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, "jessishank")
	mockRC.EXPECT().GetCred(mockStore, pb.SubCredType_BITBUCKET, ident, "jessishank", false).Return(nil, errors.New("womp womp")).Times(1)

	hhc.RepoPush(resp, req, pb.SubCredType_BITBUCKET)
	result := resp.Result()
	if result.StatusCode != http.StatusInternalServerError {
		t.Error("couldn't get vcs, should return 500")
	}
	defer result.Body.Close()
	bits, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(bits, []byte("womp womp")) {
		t.Error("did not bubble up erorr from vcs error, instead error is: " + string(bits))
	}


}

func TestRepoPush_badTranslate(t *testing.T) {
	badFp := filepath.Join("test-fixtures", "useless-json.json")
	bad, err := os.Open(badFp)
	if err != nil {
		t.Error(err)
		return
	}
	req := httptest.NewRequest("POST", "/bitbucket", bad)
	resp := httptest.NewRecorder()
	hhc, _, _, _, _, _, _ := createMockedHHC(t)
	hhc.RepoPush(resp, req, pb.SubCredType_BITBUCKET)
	result := resp.Result()
	if result.StatusCode != http.StatusBadRequest {
		t.Error("this shoudlf ail ")
	}
	hhc.PullRequest(resp, req, pb.SubCredType_BITBUCKET)
	result = resp.Result()
	if result.StatusCode != http.StatusBadRequest {
		t.Error("this shoudlf ail ")
	}
}


func TestRepoPush_werkerFailed(t *testing.T) {
	failedToTellWerker.Set(0)
	twoCommitPushFp := filepath.Join("test-fixtures", "two_commit_push.json")

	twoCommitPush, err := os.Open(twoCommitPushFp)
	if err != nil {
		t.Error(err)
	}
	req := httptest.NewRequest("POST", "/bitbucket", twoCommitPush)
	resp := httptest.NewRecorder()
	req.Header.Set("X-Event-Key","repo:push")
	hhc, mockRC, _, mockStore, handler, teller, _ := createMockedHHC(t)
	ident, _ := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, "jessishank")
	mockRC.EXPECT().GetCred(mockStore, pb.SubCredType_BITBUCKET, ident, "jessishank", false).Return(&pb.VCSCreds{AcctName:"jessishank", ClientSecret:"xxx", Identifier: ident}, nil).Times(1)
	push := &pb.Push{
		Repo: mytestOcy,
		User: jessiOwner,
		Branch: "here_we_go",
		HeadCommit: &pb.Commit{
			Hash: "4f59e57d8878daabcbc6af6bd374929b9fe03568",
			Message: "adding test data\n",
			Date: &timestamp.Timestamp{Seconds:1537998853},
			Author: jessiUser,
		},
		PreviousHeadCommit: &pb.Commit{
			Hash: "906a8969084b9ab0d8ce1056d3e046629eb2d6f9",
			Message: "adding one more change after submitting pr!!! woooahh nelly!\n",
			Date: &timestamp.Timestamp{Seconds:1537998072},
			Author: jessiUser,
		},
		Commits: []*pb.Commit{
			{
				Hash:    "4f59e57d8878daabcbc6af6bd374929b9fe03568",
				Message: "adding test data\n",
				Date: &timestamp.Timestamp{Seconds:1537998853},
				Author:  jessiUser,
			},
			{
				Hash:    "48bd5377046bdd5bcce847c50b3bce90e22bd66d",
				Message: "another one!\n",
				Author:  jessiUser,
				Date: &timestamp.Timestamp{Seconds:1537998814},
			},
		},

	}

	teller.EXPECT().TellWerker(push, hhc.Signaler, handler, "token", false, pb.SignaledBy_PUSH).Return(errors.New("wompy wompy no build 4u")).Times(1)
	hhc.HandleBBEvent(resp, req)

	metrik := &dto.Metric{}
	failedToTellWerker.Write(metrik)
	if *metrik.Counter.Value != 1 {
		t.Error("failed to tell werker, counter for this metric shoudl bump up")
	}
}

func TestHookHandlerContext_PullRequest_happy(t *testing.T) {
	prFp := filepath.Join("test-fixtures", "pr_create.json")
	prcreate, err := os.Open(prFp)
	if err != nil {
		t.Error(err)
		return
	}
	req := httptest.NewRequest("POST", "/bitbucket", prcreate)
	resp := httptest.NewRecorder()
	req.Header.Set("X-Event-Key","pullrequest:updated")
	hhc, mockRC, _, mockStore, handler, _, teller := createMockedHHC(t)
	ident, _ := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, "jessishank")
	mockRC.EXPECT().GetCred(mockStore, pb.SubCredType_BITBUCKET, ident, "jessishank", false).Return(&pb.VCSCreds{AcctName:"jessishank", ClientSecret:"xxx", Identifier: ident}, nil).Times(1)

	pr := &pb.PullRequest{
		Description: "* er my gerd\r\n* useless garbage\r\n* i wonder if this will break d3 build\r\n* just going to do one commit this time",
		Urls: &pb.PrUrls{
			Commits: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/commits",
			Comments: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/comments",
			Statuses: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/statuses",
			Decline: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/decline",
			Approve: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/approve",
			Merge: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/merge",
		},
		Title: "Here we go",
		Source: &pb.HeadData{
			Branch: "here_we_go",
			Hash: "7e4a3761b5a5",
			Repo: mytestOcy,
		},
		Destination: &pb.HeadData{
			Branch: "master",
			Hash: "8ca962b8612c",
			Repo: mytestOcy,
		},
		Id: 4,
	}
	prData := &pb.PrWerkerData{
		Urls: pr.Urls,
		PrId: "4",
	}
	teller.EXPECT().TellWerker(pr, prData, hhc.Signaler, handler, "token", false, pb.SignaledBy_PULL_REQUEST)
	hhc.HandleBBEvent(resp, req)
}


func TestHookHandlerContext_PullRequest_badwerker(t *testing.T) {
	prFp := filepath.Join("test-fixtures", "pr_create.json")
	prcreate, err := os.Open(prFp)
	if err != nil {
		t.Error(err)
		return
	}
	failedToTellWerker.Set(0)
	req := httptest.NewRequest("POST", "/bitbucket", prcreate)
	resp := httptest.NewRecorder()
	hhc, mockRC, _, mockStore, handler, _, teller := createMockedHHC(t)
	ident, _ := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, "jessishank")
	mockRC.EXPECT().GetCred(mockStore, pb.SubCredType_BITBUCKET, ident, "jessishank", false).Return(&pb.VCSCreds{AcctName:"jessishank", ClientSecret:"xxx", Identifier: ident}, nil).Times(1)

	pr := &pb.PullRequest{
		Description: "* er my gerd\r\n* useless garbage\r\n* i wonder if this will break d3 build\r\n* just going to do one commit this time",
		Urls: &pb.PrUrls{
			Commits: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/commits",
			Comments: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/comments",
			Statuses: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/statuses",
			Decline: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/decline",
			Approve: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/approve",
			Merge: "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/4/merge",
		},
		Title: "Here we go",
		Source: &pb.HeadData{
			Branch: "here_we_go",
			Hash: "7e4a3761b5a5",
			Repo: mytestOcy,
		},
		Destination: &pb.HeadData{
			Branch: "master",
			Hash: "8ca962b8612c",
			Repo: mytestOcy,
		},
		Id: 4,
	}
	prData := &pb.PrWerkerData{
		Urls: pr.Urls,
		PrId: "4",
	}
	teller.EXPECT().TellWerker(pr, prData, hhc.Signaler, handler, "token", false, pb.SignaledBy_PULL_REQUEST).Return(errors.New("nop nop nop nop")).Times(1)
	hhc.PullRequest(resp, req, pb.SubCredType_BITBUCKET)
	metrik := &dto.Metric{}
	failedToTellWerker.Write(metrik)

	if *metrik.Counter.Value != 1 {
		t.Error("failed to tell werker, counter for this metric shoudl bump up")
	}
}

func TestHookHandlerContext_HandleBBEvent_incompat(t *testing.T) {
	req := httptest.NewRequest("POST", "/bitbucket", nil)
	resp := httptest.NewRecorder()
	req.Header.Set("X-Event-Key","incompatible:updated")
	hhc, _, _, _, _, _, _ := createMockedHHC(t)
	hhc.HandleBBEvent(resp, req)
	result := resp.Result()
	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Error("incompatible X-Event-Key type, should return error")
	}
}
