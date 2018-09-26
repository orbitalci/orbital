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

	"github.com/shankj3/ocelot/models/pb"
)

func TestRepoPush(t *testing.T) {
	twoCommitPushFp := filepath.Join("test-fixtures", "two_commit_push.json")

	twoCommitPush, err := os.Open(twoCommitPushFp)
	if err != nil {
		t.Error(err)
	}
	req := httptest.NewRequest("POST", "/bitbucket", twoCommitPush)
	resp := httptest.NewRecorder()
	vcsType := pb.SubCredType_BITBUCKET

}