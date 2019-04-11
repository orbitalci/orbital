package runtime

import (
	"github.com/level11consulting/ocelot/models"
	//"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/test"
)

// fixme: this test plz
//
//func TestWorkerMsgHandler_WatchForResults(t *testing.T) {
//	var watchdata = []struct {
//		name string
//		hash string
//		chanData []byte
//	}{
//		{"ice","hash hash baby", []byte("ice ice baby")},
//		{"bean", "pinto bean", []byte("black bean")},
//		{"fruit", "jackfruit", []byte("idk.. whats like a jackfruit but *not* a jackfruit?")},
//	}
//	wmh := testGetWorkerMsgHandler(t, "test wfr")
//	for _, wd := range watchdata {
//		wd := wd
//		t.Run(wd.name, func(t *testing.T){
//			go wmh.WatchForResults(wd.hash, 1)
//			go func(){
//				wmh.infochan <- wd.chanData
//			}()
//			trans := <- wmh.StreamChan
//			info := <- trans.InfoChan
//			//t.Log("recieved")
//			if bytes.Compare(info, wd.chanData) != 0 {
//				t.Error(test.StrFormatErrors("info channel response", string(wd.chanData), string(info)))
//			}
//			if wd.hash != trans.Hash {
//				t.Error(test.StrFormatErrors("git hash", wd.hash, trans.Hash))
//			}
//		})
//	}
//}

type dummyBuildStage struct {
	details []*models.StageResult
	fail    bool
}

func (dbs *dummyBuildStage) AddStageDetail(stageResult *models.StageResult) error {
	if dbs.fail {
		return errors.New("i am failing as promised")
	}
	dbs.details = append(dbs.details, stageResult)
	return nil
}

func (dbs *dummyBuildStage) RetrieveStageDetail(buildId int64) ([]models.StageResult, error) {
	var srs []models.StageResult
	for _, i := range dbs.details {
		srs = append(srs, *i)
	}
	return srs, nil
}

func Test_handleTriggers(t *testing.T) {
	var triggerData = []struct {
		task        *pb.WerkerTask
		shouldSkip  bool
		store       *dummyBuildStage
		shouldError bool
	}{
		{&pb.WerkerTask{Branch: "boogaloo"}, true, &dummyBuildStage{details: []*models.StageResult{}}, false},
		{&pb.WerkerTask{Branch: "alks;djf"}, true, &dummyBuildStage{details: []*models.StageResult{}, fail: true}, true},
		{&pb.WerkerTask{Branch: "vibranium"}, false, &dummyBuildStage{details: []*models.StageResult{}}, false},
	}
	triggers := &pb.Triggers{Branches: []string{"apple", "banana", "quartz", "vibranium"}}
	stage := &pb.Stage{Env: []string{}, Script: []string{"echo suuuup yooo"}, Name: "testing_triggers", Trigger: triggers}

	for ind, wd := range triggerData {
		t.Run(fmt.Sprintf("%d-trigger", ind), func(t *testing.T) {
			shouldSkip, err := handleTriggers(wd.task, wd.store, stage)
			if err != nil && !wd.shouldError {
				t.Fatal(err)
			}
			if wd.shouldError && err == nil {
				t.Error("handleTriggers should have errored, didn't")
			}
			if wd.shouldSkip != shouldSkip {
				t.Logf("branch %s | shouldSkip %v | shouldError %v", wd.task.Branch, wd.shouldSkip, wd.shouldError)
				t.Error(test.GenericStrFormatErrors("should skip", wd.shouldSkip, shouldSkip))
			}
		})
	}
}
