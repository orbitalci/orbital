package recovery

import (
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
	"time"
)
//go:generate stringer -type=Interrupt
type Interrupt int

const (
	Signal Interrupt = iota
	Panic
)

type Recovery struct {
	RemoteConfig    cred.CVRemoteConfig
	WerkerUuid		uuid.UUID
	// todo: take starttime out, its irrelevant now that it is stored in consul.
	StartTime       time.Time
	CurrentStage	string
	BuildId			int64
}

func NewRecovery(rc cred.CVRemoteConfig, uid uuid.UUID) *Recovery{
	return &Recovery{RemoteConfig: rc, WerkerUuid: uid}
}

// Reset will set the build stage for the runtime of the hash, and it will add a start time.
func (r *Recovery) Reset(newStage string, hash string) error {
	consul := r.RemoteConfig.GetConsul()
	r.StartTime = time.Now()
	r.CurrentStage = newStage
	err := brt.RegisterBuildStage(consul, r.WerkerUuid.String(), hash, newStage)
	if err != nil {
		return err
	}
	err = brt.RegisterStageStartTime(consul, r.WerkerUuid.String(), hash, r.StartTime)
	return err
}

// StoreInterrupt will look up in consul all of the associated active builds with the werker and their current
// runtime state. It will then save each build's current stage details with an
func (r *Recovery) StoreInterrupt(typ Interrupt) {
	store, err := r.RemoteConfig.GetOcelotStorage()
	if err != nil {
		log.IncludeErrField(err).Error("unable to get storage when panic occured")
		return
	}
	consul := r.RemoteConfig.GetConsul()
	hrts, err := brt.GetHashRuntimesByWerker(consul, r.WerkerUuid.String())
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


// Cleanup gets all the docker uuids running according to this werker id and attempts to kill and remove the associated containers.
//   It also looks up all active builds associated with the werker id and clears them out of consul before finally deregistering itself as a werker in consul.
func (r *Recovery) Cleanup() {
	consul := r.RemoteConfig.GetConsul()
	uuids, err := brt.GetDockerUuidsByWerkerId(consul, r.WerkerUuid.String())
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
			log.Log().WithField("dockerId", uid).Info("killed container, now removing")
			if err := cli.ContainerRemove(ctx, uid, types.ContainerRemoveOptions{}); err != nil {
				log.IncludeErrField(err).WithField("dockerId", uid).Error("could not rm container")
			} else {
				log.Log().WithField("dockerId", uid).Info("removed container")
			}
		}
	}
	cli.Close()
	hashes, err := brt.GetWerkerActiveBuilds(consul, r.WerkerUuid.String())
	if err != nil {
		log.IncludeErrField(err).Error("could not get active builds for werker")
		return
	}
	for _, hash := range hashes {
		if err := brt.Delete(consul, hash); err != nil {
			log.IncludeErrField(err).WithField("gitHash", hash).Error("could not delete out of consul for build")
		} else {
			log.Log().WithField("gitHash", hash).Info("successfully delete git hashes out of build runtime consul")
		}
	}
	if err := brt.UnRegister(consul, r.WerkerUuid.String()); err != nil {
		log.IncludeErrField(err).WithField("werkerId", r.WerkerUuid.String()).Error("unable to remove werker location register out of consul.")
	} else {
		log.Log().WithField("werkerId", r.WerkerUuid.String()).Info("successfully unregistered")
	}

}

func (r *Recovery) MakeItSoDed(finish chan int) {
	if rec := recover(); rec != nil {
		defer os.Exit(1)
		fmt.Println(rec)
		log.Log().WithField("stack", string(debug.Stack())).Error("recovering from panic")
		r.StoreInterrupt(Panic)
		r.Cleanup()
		finish <- 1
	}
}
//
//func (r *Recovery) WerkerDed() {
//	if rec := recover(); rec != nil {
//		log.IncludeErrField(errors.New(string(debug.Stack()))).Error("recovering from panic")
//		r.Cleanup()
//	}
//
//}