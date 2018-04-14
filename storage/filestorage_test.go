package storage

import (
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"os"
	"testing"
	"time"
)


var fileStorage = []struct {
	hash string
	starttime time.Time
	account string
	repo string
	branch string
	failed bool
	duration float64
	id int64
	output string
}{
	{"1238ejs7", time.Now(), "jessi", "toast", "better_branch", false, 19238.834, 0, "apoweiuraoijdncvklsdixgyuaw;elkkmafs8ery239283490"},
	{"alsdkurnv", time.Now().Add(time.Second*25), "scooby", "snacks", "mystic", false, 129.87238, 0, "a;lswe39mnfxco985m.ncxzilo"},
	{"123cc34", time.Now().Add(time.Second*400), "brunswick", "york", "jersey", false, 689.3128, 0, "38a.cxv89uew,.mzxkl82!!!!!"},
}

func TestFileBuildStorage_BigOlTestBoi(t *testing.T) {
	savedirec := "./test-fixtures/big-test-boi"
	defer os.RemoveAll(savedirec)
	fbs := NewFileBuildStorage(savedirec)
	for _, fs := range fileStorage {
		//t.Run(string(ind), func(t *testing.T){
			summary := &models.BuildSummary{
				Hash: fs.hash,
				BuildTime: fs.starttime,
				Account: fs.account,
				Repo: fs.repo,
				Branch: fs.branch,
			}
			id, err := fbs.AddSumStart(fs.hash, fs.account, fs.repo, fs.branch)
			if err != nil {
				t.Error("should not have errored. error: ", err.Error())
				return
			}
			summary.BuildId = id
			sums, err := fbs.RetrieveSum(fs.hash)
			if err != nil {
				t.Error("retrieve should not have errored. error: ", err.Error())
				return
			}
			if !summary.Equals(&sums[0]) {
				t.Errorf("retrieved summary should be equal to %v, it is %v", summary, sums[0])
			}
			if err := fbs.UpdateSum(fs.failed, fs.duration, id); err != nil {
				t.Error("should have updated sum appropriately. error: ", err.Error())
				return
			}
			summary.Failed = fs.failed
			summary.BuildDuration = fs.duration
			sums, err = fbs.RetrieveSum(fs.hash)
			if !summary.Equals(&sums[0]) {
				t.Errorf("retrieved summary should be equal to %v, it is %v", summary, sums[0])
			}
			// now for testing build output
			out := &models.BuildOutput{
				BuildId: id,
				Output: []byte(fs.output),
			}
			if err := fbs.AddOut(out); err != nil {
				t.Error("should have added build output appropriatedly. error: ", err.Error())
			}
			retrieved, err := fbs.RetrieveOut(id)
			if err != nil {
				t.Error("should have successfully retrieved build output. error: ", err.Error())
			}
			if !out.Equals(&retrieved) {
				t.Errorf("retrieved output should be equal to %v, it is %v", out, &retrieved)
			}
		//}
	}

}