package valet

import (
	"fmt"
	"testing"
	"time"

	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/test"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/testutil"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common"
	cred "github.com/shankj3/ocelot/common/credentials"
	util "github.com/shankj3/ocelot/common/testutil"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/storage"
)

type recoveryCVRemoteConfig struct {
	cred.CVRemoteConfig
	consul  *consul.Consulet
	storage storage.OcelotStorage
}

func (r *recoveryCVRemoteConfig) GetConsul() *consul.Consulet {
	return r.consul
}

func (r *recoveryCVRemoteConfig) GetOcelotStorage() (storage.OcelotStorage, error) {
	return r.storage, nil
}

func addHashRuntimeData(t *testing.T, serv *testutil.TestServer, werkerId string, hash string, id int64) *build.HashRuntime {
	hrt := &build.HashRuntime{
		DockerUuid:   "here-is-my-uuid",
		BuildId:      id,
		CurrentStage: "test",
		StageStart:   time.Now(),
	}
	serv.SetKVString(t, common.MakeBuildStartpath(werkerId, hash), fmt.Sprintf("%d", hrt.StageStart.Unix()))
	serv.SetKVString(t, common.MakeDockerUuidPath(werkerId, hash), hrt.DockerUuid)
	serv.SetKVString(t, common.MakeBuildStagePath(werkerId, hash), hrt.CurrentStage)
	serv.SetKVString(t, common.MakeBuildSummaryIdPath(werkerId, hash), fmt.Sprintf("%d", hrt.BuildId))
	return hrt
}

func TestRecovery(t *testing.T) {
	// for now
	t.Skip("fix the stupid postgres bullshit")
	hash := "hahsyhashahs"
	cleanup, pw, port := storage.CreateTestPgDatabase(t)
	defer cleanup(t)
	store := storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	id, err := store.AddSumStart(hash, "test", "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	remoteConf := &recoveryCVRemoteConfig{consul: consu, storage: store}
	uid := uuid.New()
	rcvr := NewValet(remoteConf, uid, models.Docker, store, &models.SSHFacts{})
	RegisterStartedBuild(consu, uid.String(), hash)
	err = rcvr.Reset("START", hash)
	if err != nil {
		t.Fatal(err)
	}
	stage := serv.GetKVString(t, common.MakeBuildStagePath(uid.String(), hash))
	if stage != "START" {
		t.Error(test.StrFormatErrors("stage", "START", stage))
	}
	// test hash runtimes store failure
	hrt := addHashRuntimeData(t, serv, uid.String(), hash, id)
	rcvr.StoreInterrupt(Panic)
	stages, err := store.RetrieveStageDetail(hrt.BuildId)
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 1 {
		t.Error("should only be one stage")
	}
	if len(stages) == 0 {
		t.Fatal("couldn't retrieve any stages")
	}
	if stages[0].Error != "An interrupt of type Panic occurred!" {
		t.Error(test.StrFormatErrors("error message", "A panic occured!", stages[0].Error))
	}
	if stages[0].Stage != hrt.CurrentStage {
		t.Error(test.StrFormatErrors("stage", hrt.CurrentStage, stages[0].Stage))
	}
	summary, err := store.RetrieveSumByBuildId(hrt.BuildId)
	if err != nil {
		t.Fatal(err)
	}
	if summary.BuildDuration < 0 {
		t.Error("the build summary build duration should have been updated to be greater than zero when StoreInterrupt is called.")
	}
	serv.SetKVString(t, common.MakeDockerUuidPath(uid.String(), hash), hrt.DockerUuid)
	serv.SetKVString(t, common.MakeWerkerIpPath(uid.String()), "localheist")
	rcvr.RemoveAllTrace()
	// check all paths have been removed
	//
	// ci/builds/<werkerId>/<hash>
	pairs, err := consu.GetKeyValues(common.MakeBuildPath(uid.String(), hash))
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 0 {
		t.Error("everything under the build path prefix should be deleted after cleanup.")
	}
	// ci/werker_location/<werkerId>
	pairs, err = consu.GetKeyValues(common.MakeWerkerLocPath(uid.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 0 {
		t.Error("everything under the werker loc path prefix should be deleted after cleanup.")
	}
	// ci/werker_build_map/<hash>
	pairs, err = consu.GetKeyValues(common.MakeBuildMapPath(hash))
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 0 {
		t.Error("everything under the build map should be deleted after cleanup")
	}
}

func Test_Delete(t *testing.T) {
	werkerId := "werkerId"
	hash := "1231231231"
	dockerUuid := "12312324/81dfasd"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	serv.SetKV(t, fmt.Sprintf(common.BuildDockerUuid, werkerId, hash), []byte(dockerUuid))
	serv.SetKV(t, fmt.Sprintf(common.WerkerBuildMap, hash), []byte(werkerId))
	if err := Delete(consu, hash); err != nil {
		t.Fatal("could not delete!", err)
	}
	liveUuid, err := consu.GetKeyValue(fmt.Sprintf(common.BuildDockerUuid, werkerId, hash))
	if err != nil {
		t.Fatal("unable to connect to consu ", err.Error())
	}
	if liveUuid != nil {
		t.Error("liveUuid path should not exist after delete")
	}
	werkerIdd, err := consu.GetKeyValue(fmt.Sprintf(common.WerkerBuildMap, hash))
	if err != nil {
		t.Fatal("unable to connect to conu ", err.Error())
	}
	if werkerIdd != nil {
		t.Error("werkerId path should not exist after delete")
	}

}
