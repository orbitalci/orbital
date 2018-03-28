package valet

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/go-til/log"
	brt "bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"os"
	"runtime/debug"
	"sync"
	"time"
)
//go:generate stringer -type=Interrupt
type Interrupt int

const (
	Signal Interrupt = iota
	Panic
)

type Valet struct {
	RemoteConfig    cred.CVRemoteConfig
	WerkerUuid		uuid.UUID
	doneChannels    map[string]chan int
	sync.Mutex

}

func NewValet(rc cred.CVRemoteConfig, uid uuid.UUID) *Valet {
	return &Valet{RemoteConfig: rc, WerkerUuid: uid, doneChannels: make(map[string]chan int)}
}


// Reset will set the build stage for the runtime of the hash, and it will add a start time.
func (v *Valet) Reset(newStage string, hash string) error {
	consulet := v.RemoteConfig.GetConsul()
	err := brt.RegisterBuildStage(consulet, v.WerkerUuid.String(), hash, newStage)
	if err != nil {
		return err
	}
	err = brt.RegisterStageStartTime(consulet, v.WerkerUuid.String(), hash, time.Now())
	return err
}

// StoreInterrupt will look up in consul all of the associated active builds with the werker and their current
// runtime state. It will then save each build's current stage details with an
func (v *Valet) StoreInterrupt(typ Interrupt) {
	store, err := v.RemoteConfig.GetOcelotStorage()
	if err != nil {
		log.IncludeErrField(err).Error("unable to get storage when panic occured")
		return
	}
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
			BuildId: hrt.BuildId,
			StageDuration: duration,
			Stage: hrt.CurrentStage,
			Error: "An interrupt of type " + typ.String() + " occurred!",
			Messages: messages,
			StartTime: hrt.StageStart,
		}
		if err := store.AddStageDetail(detail); err != nil {
			log.IncludeErrField(err).Error("couldn't store stage detail!")
		} else {
			log.Log().Info("updated stage detail")
		}
		sum, err := store.RetrieveSumByBuildId(hrt.BuildId)
		if err != nil {
			log.IncludeErrField(err).Error("could not retrieve summary for update")
		}
		fullDuration := time.Now().Sub(sum.BuildTime).Seconds()
		if err := store.UpdateSum(true, fullDuration, hrt.BuildId); err != nil {
			log.IncludeErrField(err).Error("couldn't update summary in database")
		} else {
			log.Log().Info("updated summary table in database")
		}
	}
}

// StartBuild will register the uuid, hash, and database id into consul, as well as update the werker_id:hash kv in consul.
//  StartBuild will panic if it cannot connect to consul. idk if this is the
func (v *Valet) StartBuild(consulet *consul.Consulet, hash string, id int64) error {
	var err error
	if err = brt.RegisterBuildSummaryId(consulet, v.WerkerUuid.String(), hash, id); err != nil {
		log.IncludeErrField(err).Error("could not register build summary id into consul! huge deal!")
		return err
	}

	if err = brt.RegisterStartedBuild(consulet, v.WerkerUuid.String(), hash); err != nil {
		log.IncludeErrField(err).Error("couldn't register build")
		return err
	}
	return nil
}

// Cleanup gets all the docker uuids running according to this werker id and attempts to kill and remove the associated containers.
//   It also looks up all active builds associated with the werker id and clears them out of consul before finally deregistering itself as a werker in consul.
func (v *Valet) Cleanup() {
	consulet := v.RemoteConfig.GetConsul()
	uuids, err := brt.GetDockerUuidsByWerkerId(consulet, v.WerkerUuid.String())
	if err != nil {
		log.IncludeErrField(err).Error("unable to get docker uuids? is nothing sacred?!")
		return
	}
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.IncludeErrField(err).Error("unable to get docker client?? ")
		return
	}
	for _, uid := range uuids {
		if err := cli.ContainerKill(ctx, uid, "SIGKILL"); err != nil {
			log.IncludeErrField(err).WithField("dockerId", uid).Error("could not kill container")
		} else {
			log.Log().WithField("dockerId", uid).Info("killed container")
		}
		// even if ther is an error with containerKill, it might be from the container already exiting (ie bad ocelot.yml). so still try to remove.
		log.Log().WithField("dockerId", uid).Info("removing")
		if err := cli.ContainerRemove(ctx, uid, types.ContainerRemoveOptions{}); err != nil {
			log.IncludeErrField(err).WithField("dockerId", uid).Error("could not rm container")
		} else {
			log.Log().WithField("dockerId", uid).Info("removed container")
		}
	}

	cli.Close()
	hashes, err := brt.GetWerkerActiveBuilds(consulet, v.WerkerUuid.String())
	if err != nil {
		log.IncludeErrField(err).Error("could not get active builds for werker")
		return
	}
	for _, hash := range hashes {
		if err := brt.Delete(consulet, hash); err != nil {
			log.IncludeErrField(err).WithField("gitHash", hash).Error("could not delete out of consul for build")
		} else {
			log.Log().WithField("gitHash", hash).Info("successfully delete git hashes out of build runtime consul")
		}
	}
	if err := brt.UnRegister(consulet, v.WerkerUuid.String()); err != nil {
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
		v.Cleanup()
		finish <- 1
	}
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
	done, ok := v.doneChannels[hash]
	if !ok {
		log.Log().WithField("hash", hash).Warning("fyi! hash wasn't found in done channel map!")
	} else {
		// so i took this out of UnmarshalAndProcess and i don't know if its the best move..
		done <- 1
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
	v.Cleanup()
	os.Exit(1)
}

