package cleaner


import (
	"context"
	"github.com/docker/docker/client"
	"bitbucket.org/level11consulting/go-til/log"
	"github.com/docker/docker/api/types"
	config "bitbucket.org/level11consulting/ocelot/old/werker/config"
)

//this interface handles build cleanup
type Cleaner interface {

	//Cleanup performs build cleanup functions. If an optional logout channel is passed, logs will be sent over the channel
	Cleanup(ctx context.Context, id string, logout chan []byte) error
}

//returns a new cleaner interface
func GetNewCleaner(werkerType config.WerkType) Cleaner {
	switch werkerType {
	case config.Docker:
		return &DockerCleaner{}
	case config.Kubernetes:
		return &K8Cleaner{}
	default:
		return &DockerCleaner{}
	}
	return nil
}

type DockerCleaner struct {}

func (d *DockerCleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	cli, err := client.NewEnvClient()
	defer cli.Close()
	if err != nil {
		log.IncludeErrField(err).Error("unable to get docker client?? ")
		return err
	}

	if err := cli.ContainerKill(ctx, id, "SIGKILL"); err != nil {
		if err == context.Canceled && logout != nil {
			logout <- []byte("//////////REDRUM////////REDRUM////////REDRUM/////////")
		}
		log.IncludeErrField(err).WithField("containerId", id).Error("couldn't kill")
	} else {
		log.Log().WithField("dockerId", id).Info("killed container")
	}

	// even if ther is an error with containerKill, it might be from the container already exiting (ie bad ocelot.yml). so still try to remove.
	log.Log().WithField("dockerId", id).Info("removing")
	if err := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		log.IncludeErrField(err).WithField("dockerId", id).Error("could not rm container")
		return err
	} else {
		log.Log().WithField("dockerId", id).Info("removed container")
	}
	return nil
}

type K8Cleaner struct {}

func (k *K8Cleaner) Cleanup(ctx context.Context, id string, logout chan []byte) error {
	//TODO: implement this when the time comes
	return nil
}