package buildmonitor

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	c "github.com/level11consulting/ocelot/build/cleaner"
	"github.com/level11consulting/ocelot/client/runtime"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/storage"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/log"

	consulkv "github.com/level11consulting/ocelot/server/config/consul"

	"github.com/google/uuid"
)

//go:generate stringer -type=Interrupt
type Interrupt int

const (
	Signal Interrupt = iota
	Panic
)

// BuildMonitor is the overseer of builds. It handles registration of when the build is started, what stage it is actively on,
// when to close the channel that signifies to nsqpb to stop refreshing the status of the message
type BuildMonitor struct {
	RemoteConfig config.CVRemoteConfig
	store        storage.OcelotStorage
	WerkerUuid   uuid.UUID
	doneChannels map[string]chan int
	*BuildReaper
	sync.Mutex
	c.Cleaner
}

func NewBuildMonitor(rc config.CVRemoteConfig, uid uuid.UUID, werkerType models.WerkType, store storage.OcelotStorage, facts *models.SSHFacts) *BuildMonitor {
	buildmonitor := &BuildMonitor{RemoteConfig: rc, WerkerUuid: uid, doneChannels: make(map[string]chan int), store: store}
	buildmonitor.Cleaner = c.GetNewCleaner(werkerType, facts)
	buildmonitor.BuildReaper = NewBuildReaper()
	return buildmonitor 
}

// Reset will set the build stage for the runtime of the hash, and it will add a start time.
func (bm *BuildMonitor) Reset(newStage string, hash string) error {
	consulet := bm.RemoteConfig.GetConsul()
	err := RegisterBuildStage(consulet, bm.WerkerUuid.String(), hash, newStage)
	if err != nil {
		return err
	}
	err = RegisterStageStartTime(consulet, bm.WerkerUuid.String(), hash, time.Now())
	return err
}

// StoreInterrupt will look up in consul all of the associated active builds with the werker and their current
// runtime state. It will then save each build's current stage details with an
func (bm *BuildMonitor) StoreInterrupt(typ Interrupt) {
	consulet := bm.RemoteConfig.GetConsul()
	hrts, err := runtime.GetHashRuntimesByWerker(consulet, bm.WerkerUuid.String())
	if err != nil {
		log.IncludeErrField(err).Error("unable to get hash runtimes")
		return
	}
	var messages []string
	switch typ {
	case Panic:
		messages = append(messages, string(debug.Stack()))
	case Signal:
		messages = append(messages, "The werker was interrupted with a signal")
	}

	for _, hrt := range hrts {
		duration := time.Now().Sub(hrt.StageStart).Seconds()
		detail := &models.StageResult{
			BuildId:       hrt.BuildId,
			StageDuration: duration,
			Stage:         hrt.CurrentStage,
			Error:         "An interrupt of type " + typ.String() + " occurred!",
			Messages:      messages,
			StartTime:     hrt.StageStart,
		}
		if err := bm.store.AddStageDetail(detail); err != nil {
			log.IncludeErrField(err).Error("couldn't store stage detail!")
		} else {
			log.Log().Info("updated stage detail")
		}
		sum, err := bm.store.RetrieveSumByBuildId(hrt.BuildId)
		if err != nil {
			log.IncludeErrField(err).Error("could not retrieve summary for update")
		}
		buildTime := time.Unix(sum.BuildTime.Seconds, int64(sum.BuildTime.Nanos))
		fullDuration := time.Since(buildTime).Seconds()
		if err := bm.store.UpdateSum(true, fullDuration, hrt.BuildId); err != nil {
			log.IncludeErrField(err).Error("couldn't update summary in database")
		} else {
			log.Log().Info("updated summary table in database")
		}
	}
}

// StartBuild will register the uuid, hash, and database id into consul, as well as update the werker_id:hash kv in consul.
func (bm *BuildMonitor) StartBuild(consulet consul.Consuletty, hash string, id int64) error {
	var err error
	if err = RegisterBuildSummaryId(consulet, bm.WerkerUuid.String(), hash, id); err != nil {
		log.IncludeErrField(err).Error("could not register build summary id into consul! huge deal!")
		return err
	}

	if err = RegisterStartedBuild(consulet, bm.WerkerUuid.String(), hash); err != nil {
		log.IncludeErrField(err).Error("couldn't register build")
		return err
	}
	return nil
}

// RemoveAllTrace gets all the docker uuids running according to this werker id and attempts to kill and remove the associated containers.
//   It also looks up all active builds associated with the werker id and clears them out of consul before finally deregistering itself as a werker in consul.
func (bm *BuildMonitor) RemoveAllTrace() {
	consulet := bm.RemoteConfig.GetConsul()
	uuids, err := runtime.GetDockerUuidsByWerkerId(consulet, bm.WerkerUuid.String())
	if err != nil {
		log.IncludeErrField(err).Error("unable to get docker uuids? is nothing sacred?!")
		return
	}
	ctx := context.Background()
	for _, uid := range uuids {
		bm.Cleanup(ctx, uid, nil)
	}
	log.Log().Info("cleaned up docker remnants")
	hashes, err := runtime.GetWerkerActiveBuilds(consulet, bm.WerkerUuid.String())
	if err != nil {
		log.IncludeErrField(err).Error("could not get active builds for werker")
		return
	}
	log.Log().Info("deleting hashes associated with this werker out of consul.")
	for _, hash := range hashes {
		if err := Delete(consulet, hash); err != nil {
			log.IncludeErrField(err).WithField("gitHash", hash).Error("could not delete out of consul for build")
		} else {
			log.Log().WithField("gitHash", hash).Info("successfully delete git hashes out of build runtime consul")
		}
	}
	log.Log().Info("unregister-ing myself with consul as a werker")
	if err := UnRegister(consulet, bm.WerkerUuid.String()); err != nil {
		log.IncludeErrField(err).WithField("werkerId", bm.WerkerUuid.String()).Error("unable to remove werker location register out of consul.")
	} else {
		log.Log().WithField("werkerId", bm.WerkerUuid.String()).Info("successfully unregistered")
	}
}

//MakeItSoDed is a defer recovery function for when the werker has panicked
func (bm *BuildMonitor) MakeItSoDed(finish chan int) {
	if rec := recover(); rec != nil {
		defer os.Exit(1)
		fmt.Println(rec)
		log.Log().WithField("stack", string(debug.Stack())).Error("recovering from panic")
		bm.StoreInterrupt(Panic)
		bm.RemoveAllTrace()
	}
	finish <- 1
	log.Log().Error("shutting down")
	time.Sleep(2 * time.Second)
	os.Exit(1)
}

//RegisterDoneChan will write the nsq done channel to the in-memory map associated with this buildmonitor
func (bm *BuildMonitor) RegisterDoneChan(hash string, done chan int) {
	bm.Lock()
	defer bm.Unlock()
	_, ok := bm.doneChannels[hash]
	if ok {
		// not sure if this would happen ever
		log.Log().WithField("hash", hash).Warning("fyi! overwriting hash done channel!")
	}
	bm.doneChannels[hash] = done
}

//UnregisterDoneChan will delete the done channel associated with the given hash out of the in-memory map associated with this buildmonitor
func (bm *BuildMonitor) UnregisterDoneChan(hash string) {
	bm.Lock()
	defer bm.Unlock()
	_, ok := bm.doneChannels[hash]
	if !ok {
		log.Log().WithField("hash", hash).Warning("fyi! hash wasn't found in done channel map!")
	}
	delete(bm.doneChannels, hash)

}

//CallDoneForEverything will iterate over every entry in the in-memory done map and and call "done" on it (ie send an integer over the channel)
func (bm *BuildMonitor) CallDoneForEverything() {
	// this will add to every done channel in its doneChannels map, triggering the nsqpb library to call Finish()
	for _, done := range bm.doneChannels {
		done <- 1
	}
}

//SignalRecvDed is responsible for closing out all active work when a werker has recieved a signal (SIGKILL, etc). It will store in the database
// that it has been interrupted for every acssociated active build, it will call done for every nsq active connection, and it will delete all its
// entries out of the database, then it will finally exit w/ code 0
func (bm *BuildMonitor) SignalRecvDed() {
	log.Log().Info("received interrupt, cleaning up after myself...")
	bm.StoreInterrupt(Signal)
	bm.CallDoneForEverything()
	bm.RemoveAllTrace()
	os.Exit(0)
}

// Delete will remove everything related to that werker's build of the gitHash out of consul
// will delete:
// 		ci/werker_build_map/<hash>
// 		ci/builds/<werkerId>/<hash>/*
func Delete(consulete consul.Consuletty, gitHash string) (err error) {
	//paths := &Identifiers{GitHash: gitHash}
	pairPath := consulkv.MakeBuildMapPath(gitHash)
	kv, err := consulete.GetKeyValue(pairPath)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't get kv error!")
		return
	}
	if kv == nil {
		log.Log().Error("THIS PAIR SHOULD NOT BE NIL! path: " + pairPath)
		return
	}
	log.Log().WithField("gitHash", gitHash).Info("WERKERID IS: ", string(kv.Value))
	if err = consulete.RemoveValues(consulkv.MakeBuildPath(string(kv.Value), gitHash)); err != nil {
		return
	}
	err = consulete.RemoveValue(pairPath)
	return err
}
