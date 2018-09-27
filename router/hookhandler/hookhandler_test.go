package hookhandler

//import (
//	"io/ioutil"
//	"testing"
//)
//
//func TestRepoPush(t *testing.T) {
//	//repoPushData := ioutil.ReadAll("test-fixtures/")
//}

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models/mock_models"

	//"github.com/shankj3/ocelot/models/mock_models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

func createMockedHHC(t *testing.T) (*HookHandlerContext, *credentials.MockCVRemoteConfig, *nsqpb.MockProducer, *storage.MockOcelotStorage, *mock_models.MockVCSHandler, *build_signaler.MockWerkerTeller){
	ctl := gomock.NewController(t)
	handler := mock_models.NewMockVCSHandler(ctl)
	rc := credentials.NewMockCVRemoteConfig(ctl)
	produce := nsqpb.NewMockProducer(ctl)
	store := storage.NewMockOcelotStorage(ctl)
	teller := build_signaler.NewMockWerkerTeller(ctl)
	hhc := &HookHandlerContext{
		Signaler: &build_signaler.Signaler{
			RC: 	      rc,
			Producer:     produce,
			Store: 	      store,
			Deserializer: deserialize.New(),
			OcyValidator: build.GetOcelotValidator(),
		},
		testingHandler: handler,
		teller: teller,
	}
	return hhc, rc, produce, store, handler, teller
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

func TestRepoPush(t *testing.T) {
	twoCommitPushFp := filepath.Join("test-fixtures", "two_commit_push.json")

	twoCommitPush, err := os.Open(twoCommitPushFp)
	if err != nil {
		t.Error(err)
	}
	req := httptest.NewRequest("POST", "/bitbucket", twoCommitPush)
	resp := httptest.NewRecorder()

	hhc, mockRC, _, mockStore, handler, teller := createMockedHHC(t)
	ident, _ := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, "jessishank")
	mockRC.EXPECT().GetCred(mockStore, pb.SubCredType_BITBUCKET, ident, "jessishank", false).Return(&pb.VCSCreds{AcctName:"jessishank", ClientSecret:"xxx", Identifier: ident}, nil).Times(1)
	teller.EXPECT().TellWerker("4f59e57d8878daabcbc6af6bd374929b9fe03568", hhc.Signaler, "here_we_go", handler, gomock.Any(), "jessishank/mytestocy", twoPushCommits, false, pb.SignaledBy_PUSH).Times(1)
	hhc.RepoPush(resp, req, pb.SubCredType_BITBUCKET)


}