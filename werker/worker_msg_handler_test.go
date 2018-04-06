package werker

import (
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"bytes"
	"fmt"
	"testing"

	"bitbucket.org/level11consulting/go-til/test"
	pb "bitbucket.org/level11consulting/ocelot/protos"
)

func TestWorkerMsgHandler_WatchForResults(t *testing.T) {
	var watchdata = []struct {
		name string
		hash string
		chanData []byte
	}{
		{"ice","hash hash baby", []byte("ice ice baby")},
		{"bean", "pinto bean", []byte("black bean")},
		{"fruit", "jackfruit", []byte("idk.. whats like a jackfruit but *not* a jackfruit?")},
	}
	wmh := testGetWorkerMsgHandler(t, "test wfr")
	for _, wd := range watchdata {
		wd := wd
		t.Run(wd.name, func(t *testing.T){
			go wmh.WatchForResults(wd.hash, 1)
			go func(){
				wmh.infochan <- wd.chanData
			}()
			trans := <- wmh.StreamChan
			info := <- trans.InfoChan
			//t.Log("recieved")
			if bytes.Compare(info, wd.chanData) != 0 {
				t.Error(test.StrFormatErrors("info channel response", string(wd.chanData), string(info)))
			}
			if wd.hash != trans.Hash {
				t.Error(test.StrFormatErrors("git hash", wd.hash, trans.Hash))
			}
		})
	}
}


type dummyBuildStage struct {
	details []*models.StageResult
}
func (dbs *dummyBuildStage) AddStageDetail(stageResult *models.StageResult) error {
	dbs.details = append(dbs.details, stageResult)
	return nil
}

func (dbs *dummyBuildStage) RetrieveStageDetail(buildId int64) ([]models.StageResult, error){
	var srs []models.StageResult
	for _, i := range dbs.details {
		srs = append(srs, *i)
	}
	return srs, nil
}

func Test_handleTriggers(t *testing.T) {
	var triggerData = []struct {
		branch string
		shouldSkip bool
	}{
		{"boogaloo", true},
		{"alks;djf", true},
		{"vibranium", false},
	}
	triggers := &pb.Triggers{Branches: []string{"apple", "banana", "quartz", "vibranium"}}
	stage := &pb.Stage{Env: []string{}, Script: []string{"echo suuuup yooo"}, Name: "testing_triggers", Trigger: triggers}

	store := &dummyBuildStage{details:[]*models.StageResult{}}
	for ind, wd := range triggerData {
		t.Run(fmt.Sprintf("%d-trigger", ind), func(t *testing.T){
			shouldSkip, _ := handleTriggers(wd.branch, 12, store, stage)
			if wd.shouldSkip != shouldSkip {
				t.Error(test.GenericStrFormatErrors("should skip", wd.shouldSkip, shouldSkip))
			}
		})
	}
}