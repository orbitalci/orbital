package launcher

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/basher"
	valet2 "github.com/shankj3/ocelot/build/valet"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)


func getLoopbackIp(t *testing.T) string {
	var loopIp string
	switch runtime.GOOS {
	case "darwin":
		loopIp = "docker.for.mac.localhost"
	case "linux":
		loopIp = "172.17.0.1"
	default:
		t.Skip("skipping launcher related test because only supported on darwin or linux")
	}
	return loopIp
}

func getTestBasher(t *testing.T) *basher.Basher {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	loop := getLoopbackIp(t)
	bshr, err := basher.NewBasher("", "", loop, filepath.Join(dir, "test-fixtures"))
	if err != nil {
		t.Fatal(err)
	}
	return bshr

}

func getTestingLauncher(t *testing.T) (*launcher, func(t *testing.T)) {
	remoteConf, listener, testserver := credentials.TestSetupVaultAndConsul(t)
	port := 5496
	cleanup, pw := storage.CreateTestPgDatabase(t, port)
	pg := storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	uid := uuid.New()
	valet := valet2.NewValet(remoteConf, uid, models.Docker, pg, nil)
	loopIp := getLoopbackIp(t)
	facts := &models.WerkerFacts{
		Uuid: uid,
		WerkerType: models.Docker,
		LoopbackIp: loopIp,
		RegisterIP: "localhost",
		ServicePort: "9090",
		GrpcPort: "9099",
	}
	stream := make(chan *models.Transport, 1000)
	buildCtx := make(chan *models.BuildContext, 1000)
	bshr := getTestBasher(t)
	launcher := NewLauncher(facts, remoteConf, stream, buildCtx, bshr, pg, valet)
	cleanitall := func(t *testing.T){
		cleanup(t)
		credentials.TeardownVaultAndConsul(listener, testserver)
	}
	return launcher, cleanitall
}


type testStore struct {
	storage.OcelotStorage
	addedSum *pb.BuildSummary
	stageDetails []models.StageResult
}

func (t *testStore) AddSumStart(hash string, account string, repo string, branch string) (int64, error) {
	t.addedSum =&pb.BuildSummary{Hash:hash, Account:account, Repo:repo, Branch:branch, Status: pb.BuildStatus_QUEUED, BuildTime: &timestamp.Timestamp{Seconds: 0, Nanos: 0}}
	return 1, nil
}

func (t *testStore) RetrieveStageDetail(buildId int64) ([]models.StageResult, error) {
	return t.stageDetails, nil
}

func (t *testStore) AddStageDetail(stageResult *models.StageResult) error {
	t.stageDetails = append(t.stageDetails, *stageResult)
	return nil
}

func (t *testStore) UpdateSum(failed bool, duration float64, id int64) error {
	t.addedSum.Failed = failed
	if failed {
		t.addedSum.Status = pb.BuildStatus_FAILED
	} else {
		t.addedSum.Status = pb.BuildStatus_PASSED
	}
	t.addedSum.BuildId = id
	return nil
}

func (t *testStore) RetrieveSumByBuildId(buildId int64) (*pb.BuildSummary, error) {
	return t.addedSum, nil
}


func (t *testStore) RetrieveCredBySubTypeAndAcct(scredType pb.SubCredType, acctName string) ([]pb.OcyCredder, error) {
	return nil, nil
}


type fakeBuilder struct {
	failInit bool
	failSetup bool
	failExecute bool
	failExecuteIntegration bool
	failNum int
	currentnum int
	stagesRan []*pb.Stage
	taskGiven *pb.WerkerTask
	setEnvs []string
	addedEnvs []string
	uid uuid.UUID
	*basher.Basher
}

func (f *fakeBuilder) Init(ctx context.Context, hash string, logout chan []byte) *pb.Result {
	if f.failInit {
		return &pb.Result{Status:pb.StageResultVal_FAIL, Error:"i was told to fail", Messages: []string{"fail!"}, Stage:"init"}
	}
	return &pb.Result{Status:pb.StageResultVal_PASS, Messages:[]string{"passed!"}, Stage:"init"}
}

func (f *fakeBuilder) SetGlobalEnv(envs []string) {
	f.setEnvs = envs
}

func (f *fakeBuilder) AddGlobalEnvs(envs []string) {
	f.addedEnvs = append(f.addedEnvs, envs...)
}

func (f *fakeBuilder) Setup(ctx context.Context, logout chan []byte, dockerId chan string, werk *pb.WerkerTask, rc credentials.CVRemoteConfig, werkerPort string) (res *pb.Result, uid string) {
	dockerId <- "smurf"
	close(dockerId)
	f.uid = uuid.New()
	if f.failSetup {
		return &pb.Result{Status:pb.StageResultVal_FAIL, Error:"i was told to fail", Messages: []string{"fail!"}, Stage:"setup"}, f.uid.String()
	}
	return &pb.Result{Status:pb.StageResultVal_PASS, Messages: []string{"passed setup!!"}, Stage: "setup"}, f.uid.String()
}

func (f *fakeBuilder) Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result {
	if f.failExecute {
		return &pb.Result{Status:pb.StageResultVal_FAIL, Error:"i was told to fail", Messages: []string{"fail!"}, Stage:actions.Name}
	}
	f.stagesRan = append(f.stagesRan, actions)
	return &pb.Result{Status:pb.StageResultVal_PASS, Messages:[]string{"passing stage"}, Stage: actions.Name}
}

func (f *fakeBuilder) ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *build.StageUtil, logout chan []byte) *pb.Result {
	failureres := &pb.Result{Status:pb.StageResultVal_FAIL, Error:"i was told to fail", Messages: []string{"fail!"}, Stage:stgUtil.Stage}
	if f.failExecuteIntegration {
		if f.failNum != 0 {
			if f.failNum == f.currentnum {
				return failureres
			}
			f.currentnum++
		} else {
			return failureres
		}
	}
	f.stagesRan = append(f.stagesRan, stage)
	return &pb.Result{Status:pb.StageResultVal_PASS,  Messages: []string{"passssin"}, Stage:stgUtil.Stage}
}

func (f *fakeBuilder) GetContainerId() string {
	return "smurf"
}

func (f *fakeBuilder) Close() error {
	return nil
}