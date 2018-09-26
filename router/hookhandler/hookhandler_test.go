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

func createMockedHHC(t *testing.T) (*HookHandlerContext, *credentials.MockCVRemoteConfig, *nsqpb.MockProducer, *storage.MockOcelotStorage, *mock_models.MockVCSHandler){
	ctl := gomock.NewController(t)
	handler := mock_models.NewMockVCSHandler(ctl)
	rc := credentials.NewMockCVRemoteConfig(ctl)
	produce := nsqpb.NewMockProducer(ctl)
	store := storage.NewMockOcelotStorage(ctl)
	teller
	hhc := &HookHandlerContext{
		Signaler: &build_signaler.Signaler{
			RC: 	      rc,
			Producer:     produce,
			Store: 	      store,
			Deserializer: deserialize.New(),
			OcyValidator: build.GetOcelotValidator(),
		},
		testingHandler: handler,
		teller:
	}
	return hhc, rc, produce, store, handler
}

func TestRepoPush(t *testing.T) {
	twoCommitPushFp := filepath.Join("test-fixtures", "two_commit_push.json")

	twoCommitPush, err := os.Open(twoCommitPushFp)
	if err != nil {
		t.Error(err)
	}
	req := httptest.NewRequest("POST", "/bitbucket", twoCommitPush)
	resp := httptest.NewRecorder()
	vcsType := pb.SubCredType_BITBUCKET

	hhc, mockRC, mockProduce, mockStore, handler := createMockedHHC(t)
	ident, _ := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, "shankj3")
	mockRC.EXPECT().GetCred(mockStore, pb.SubCredType_BITBUCKET, ident, "shankj3", false).Return(&pb.VCSCreds{AcctName:"shankj3", ClientSecret:"xxx", Identifier: ident}, nil).Times(1)


}