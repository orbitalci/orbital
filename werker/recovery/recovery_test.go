package recovery

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/util"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/testutil"
	"testing"
	"time"
)

type recoveryCVRemoteConfig struct {
	cred.CVRemoteConfig
	consul *consul.Consulet
	storage storage.OcelotStorage
}

func (r *recoveryCVRemoteConfig) GetConsul() *consul.Consulet {
	return r.consul
}

func (r *recoveryCVRemoteConfig) GetOcelotStorage() (storage.OcelotStorage, error) {
	return r.storage, nil
}

func addHashRuntimeData(t *testing.T, serv *testutil.TestServer, werkerId string, hash string, id int64) *buildruntime.HashRuntime {
	hrt := &buildruntime.HashRuntime{
		DockerUuid: "here-is-my-uuid",
		BuildId: id,
		CurrentStage: "test",
		StageStart: time.Now(),
	}
	serv.SetKVString(t, buildruntime.MakeBuildStartpath(werkerId, hash), fmt.Sprintf("%d", hrt.StageStart.Unix()))
	serv.SetKVString(t, buildruntime.MakeDockerUuidPath(werkerId, hash), hrt.DockerUuid)
	serv.SetKVString(t, buildruntime.MakeBuildStagePath(werkerId, hash), hrt.CurrentStage)
	serv.SetKVString(t, buildruntime.MakeBuildSummaryIdPath(werkerId, hash), fmt.Sprintf("%d", hrt.BuildId))
	return hrt
}

func TestRecovery(t *testing.T) {
	// for now
	t.Skip("fix the stupid postgres bullshit")
	hash := "hahsyhashahs"
	cleanup, pw, port := storage.CreateTestPgDatabase(t)
	defer cleanup(t)
	store := storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	id, err := store.AddSumStart(hash, time.Now(), "test", "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	remoteConf := &recoveryCVRemoteConfig{consul: consu, storage: store}
	uid := uuid.New()
	rcvr := NewRecovery(remoteConf, uid)
	buildruntime.RegisterStartedBuild(consu, uid.String(), hash)
	err = rcvr.Reset("START", hash)
	if err != nil {
		t.Fatal(err)
	}
	stage := serv.GetKVString(t, buildruntime.MakeBuildStagePath(uid.String(), hash))
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
	serv.SetKVString(t, buildruntime.MakeDockerUuidPath(uid.String(), hash), hrt.DockerUuid)
	serv.SetKVString(t, buildruntime.MakeWerkerIpPath(uid.String()), "localheist")
	rcvr.Cleanup()
	// check all paths have been removed
	//
	// ci/builds/<werkerId>/<hash>
	pairs, err := consu.GetKeyValues(buildruntime.MakeBuildPath(uid.String(), hash))
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 0 {
		t.Error("everything under the build path prefix should be deleted after cleanup.")
	}
	// ci/werker_location/<werkerId>
	pairs, err = consu.GetKeyValues(buildruntime.MakeWerkerLocPath(uid.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 0 {
		t.Error("everything under the werker loc path prefix should be deleted after cleanup.")
	}
	// ci/werker_build_map/<hash>
	pairs, err = consu.GetKeyValues(buildruntime.MakeBuildMapPath(hash))
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 0 {
		t.Error("everything under the build map should be deleted after cleanup")
	}
}
