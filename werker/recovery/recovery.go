package recovery

import (
	"bitbucket.org/level11consulting/go-til/log"
	brt "bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"os"
	"runtime/debug"
	"time"
)

type Recovery struct {
	RemoteConfig    cred.CVRemoteConfig
	WerkerUuid		uuid.UUID
	StartTime       time.Time
	CurrentStage	string
	BuildId			int64
}

func NewRecovery(rc cred.CVRemoteConfig, uid uuid.UUID) *Recovery{
	return &Recovery{RemoteConfig: rc, WerkerUuid: uid}
}

func (r *Recovery) Reset(newStage string) {
	r.StartTime = time.Now()
	r.CurrentStage = newStage
}

func (r *Recovery) StoreFailure() {
	store, err := r.RemoteConfig.GetOcelotStorage()
	if err != nil {
		log.IncludeErrField(err).Error("unable to get storage when panic occured")
	}
	detail := &models.StageResult{
		BuildId:r.BuildId,
		StageDuration: float64(time.Now().Sub(r.StartTime)),
		Stage: r.CurrentStage,
		Error: "A panic occured!",
		Messages: []string{string(debug.Stack())},
		StartTime: r.StartTime,
	}
	store.AddStageDetail(detail)
}


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
			if err := cli.ContainerRemove(ctx, uid, types.ContainerRemoveOptions{}); err != nil {
				log.IncludeErrField(err).WithField("dockerId", uid).Error("could not rm container")
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
		}
	}
	if err := brt.UnRegister(consul, r.WerkerUuid.String()); err != nil {
		log.IncludeErrField(err).WithField("werkerId", r.WerkerUuid.String()).Error("unable to remove werker location register out of consul.")
	}

}

func (r *Recovery) MakeItSoDed() {
	if rec := recover(); rec != nil {
		defer os.Exit(1)
		log.IncludeErrField(errors.New(string(debug.Stack()))).Error("recovering from panic")
		r.StoreFailure()
		r.Cleanup()
		panic(string(debug.Stack()))
	}
}

func (r *Recovery) WerkerDed() {
	if rec := recover(); rec != nil {
		log.IncludeErrField(errors.New(string(debug.Stack()))).Error("recovering from panic")
		r.Cleanup()
	}

}

//func (r *Recovery) MsgDed(msg *nsq.Message) {
//	if rec := recover(); rec != nil {
//		log.IncludeErrField(errors.New(string(debug.Stack()))).Error("recovering from panic in message, re-queueing message with 2sec delay")
//		debug.PrintStack()
//		fmt.Println("requeueing message with 2sec delay")
//		msg.Requeue(2*time.Second)
//	}
//}
