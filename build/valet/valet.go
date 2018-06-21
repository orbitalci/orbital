package valet

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/log"
	brt "github.com/shankj3/ocelot/build"
	c "github.com/shankj3/ocelot/build/cleaner"
	"github.com/shankj3/ocelot/common"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/storage"

	"github.com/google/uuid"
)

//go:generate stringer -type=Interrupt
type Interrupt int

const (
	Signal Interrupt = iota
	Panic
)

// Valet is the overseer of builds. It handles registration of when the build is started, what stage it is actively on,
// when to close the channel that signifies to nsqpb to stop refreshing the status of the message
type Valet struct {
	RemoteConfig    cred.CVRemoteConfig
	store			storage.OcelotStorage
	WerkerUuid		uuid.UUID
	doneChannels    map[string]chan int
	*ContextValet
	sync.Mutex
	c.Cleaner
}

func NewValet(rc cred.CVRemoteConfig, uid uuid.UUID, werkerType models.WerkType, store storage.OcelotStorage, facts *models.SSHFacts) *Valet {
	valet := &Valet{RemoteConfig: rc, WerkerUuid: uid, doneChannels: make(map[string]chan int), store: store}
	valet.Cleaner = c.GetNewCleaner(werkerType, facts)
	valet.ContextValet = NewContextValet()
	return valet
}

// Reset will set the build stage for the runtime of the hash, and it will add a start time.
func (v *Valet) Reset(newStage string, hash string) error {
	consulet := v.RemoteConfig.GetConsul()
	err := RegisterBuildStage(consulet, v.WerkerUuid.String(), hash, newStage)
	if err != nil {
		return err
	}
	err = RegisterStageStartTime(consulet, v.WerkerUuid.String(), hash, time.Now())
	return err
}

// StoreInterrupt will look up in consul all of the associated active builds with the werker and their current
// runtime state. It will then save each build's current stage details with an
func (v *Valet) StoreInterrupt(typ Interrupt) {
	consulet := v.RemoteConfig.GetConsul()
	hrts, err := brt.GetHashRuntimesByWerker(consulet, v.WerkerUuid.String())
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
		if err := v.store.AddStageDetail(detail); err != nil {
			log.IncludeErrField(err).Error("couldn't store stage detail!")
		} else {
			log.Log().Info("updated stage detail")
		}
		sum, err := v.store.RetrieveSumByBuildId(hrt.BuildId)
		if err != nil {
			log.IncludeErrField(err).Error("could not retrieve summary for update")
		}
		buildTime := time.Unix(sum.BuildTime.Seconds, int64(sum.BuildTime.Nanos))
		fullDuration := time.Now().Sub(buildTime).Seconds()
		if err := v.store.UpdateSum(true, fullDuration, hrt.BuildId); err != nil {
			log.IncludeErrField(err).Error("couldn't update summary in database")
		} else {
			log.Log().Info("updated summary table in database")
		}
	}
}

// StartBuild will register the uuid, hash, and database id into consul, as well as update the werker_id:hash kv in consul.
func (v *Valet) StartBuild(consulet consul.Consuletty, hash string, id int64) error {
	var err error
	if err = RegisterBuildSummaryId(consulet, v.WerkerUuid.String(), hash, id); err != nil {
		log.IncludeErrField(err).Error("could not register build summary id into consul! huge deal!")
		return err
	}

	if err = RegisterStartedBuild(consulet, v.WerkerUuid.String(), hash); err != nil {
		log.IncludeErrField(err).Error("couldn't register build")
		return err
	}
	return nil
}

// RemoveAllTrace gets all the docker uuids running according to this werker id and attempts to kill and remove the associated containers.
//   It also looks up all active builds associated with the werker id and clears them out of consul before finally deregistering itself as a werker in consul.
func (v *Valet) RemoveAllTrace() {
	consulet := v.RemoteConfig.GetConsul()
	uuids, err := brt.GetDockerUuidsByWerkerId(consulet, v.WerkerUuid.String())
	if err != nil {
		log.IncludeErrField(err).Error("unable to get docker uuids? is nothing sacred?!")
		return
	}
	ctx := context.Background()
	for _, uid := range uuids {
		v.Cleanup(ctx, uid, nil)
	}
	log.Log().Info("cleaned up docker remnants")
	hashes, err := brt.GetWerkerActiveBuilds(consulet, v.WerkerUuid.String())
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
	if err := UnRegister(consulet, v.WerkerUuid.String()); err != nil {
		log.IncludeErrField(err).WithField("werkerId", v.WerkerUuid.String()).Error("unable to remove werker location register out of consul.")
	} else {
		log.Log().WithField("werkerId", v.WerkerUuid.String()).Info("successfully unregistered")
	}
}

func (v *Valet) MakeItSoDed(finish chan int) {
	if rec := recover(); rec != nil {
		defer os.Exit(1)
		fmt.Println(rec)
		log.Log().WithField("stack", string(debug.Stack())).Error("recovering from panic")
		v.StoreInterrupt(Panic)
		v.RemoveAllTrace()
	}
	finish <- 1
	log.Log().Error("shutting down")
	time.Sleep(2 * time.Second)
	os.Exit(1)
}

func (v *Valet) RegisterDoneChan(hash string, done chan int) {
	v.Lock()
	defer v.Unlock()
	_, ok := v.doneChannels[hash]
	if ok {
		// not sure if this would happen ever
		log.Log().WithField("hash", hash).Warning("fyi! overwriting hash done channel!")
	}
	v.doneChannels[hash] = done
}

func (v *Valet) UnregisterDoneChan(hash string) {
	v.Lock()
	defer v.Unlock()
	_, ok := v.doneChannels[hash]
	if !ok {
		log.Log().WithField("hash", hash).Warning("fyi! hash wasn't found in done channel map!")
	}
	delete(v.doneChannels, hash)

}

func (v *Valet) CallDoneForEverything() {
	// this will add to every done channel in its doneChannels map, triggering the nsqpb library to call Finish()
	for _, done := range v.doneChannels {
		done <- 1
	}
}

func (v *Valet) SignalRecvDed() {
	log.Log().Info("received interrupt, cleaning up after myself...")
	v.StoreInterrupt(Signal)
	v.CallDoneForEverything()
	v.RemoveAllTrace()
	os.Exit(1)
}

// Delete will remove everything related to that werker's build of the gitHash out of consul
// will delete:
// 		ci/werker_build_map/<hash>
// 		ci/builds/<werkerId>/<hash>/*
func Delete(consulete consul.Consuletty, gitHash string) (err error) {
	//paths := &Identifiers{GitHash: gitHash}
	pairPath := common.MakeBuildMapPath(gitHash)
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
	if err = consulete.RemoveValues(common.MakeBuildPath(string(kv.Value), gitHash)); err != nil {
		return
	}
	err = consulete.RemoveValue(pairPath)
	return err
}
