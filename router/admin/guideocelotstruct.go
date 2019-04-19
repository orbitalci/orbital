package admin

import (

	"github.com/level11consulting/ocelot/models"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"

	"github.com/level11consulting/ocelot/build/helpers/buildscript/validate"
	"github.com/level11consulting/ocelot/client/buildconfigvalidator"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/storage"
	"github.com/level11consulting/ocelot/router/admin/anycred"
)

//this is our grpc server, it responds to client requests
type guideOcelotServer struct {
	RemoteConfig   config.CVRemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *validate.AdminValidator
	RepoValidator  *validate.RepoValidator
	OcyValidator   *buildconfigvalidator.OcelotValidator
	Storage        storage.OcelotStorage
	Producer       nsqpb.Producer
	handler        models.VCSHandler
	hhBaseUrl      string
}

type OcelotServerAPI struct {
	DeprecatedHandler *guideOcelotServer
	anycred.AnyCredAPI // This is a hack. Revisit once stable
	BuildAPI
	AppleDevSecretAPI
	ArtifactRepoSecretAPI
	GenericSecretAPI
	KubernetesSecretAPI
	NotifierSecretAPI
	PollScheduleAPI
	RepoInterfaceAPI
	SshSecretAPI
	StatusInterfaceAPI
	VcsSecretAPI
}

func NewGuideOcelotServer(config config.CVRemoteConfig, d *deserialize.Deserializer, adminV *validate.AdminValidator, repoV *validate.RepoValidator, storage storage.OcelotStorage, hhBaseUrl string) pb.GuideOcelotServer {
	// changing to this style of instantiation cuz thread safe (idk read it on some best practices, it just looks
	// purdier to me anyway
	guideOcelotServer := &guideOcelotServer{
		OcyValidator:   buildconfigvalidator.GetOcelotValidator(),
		RemoteConfig:   config,
		Deserializer:   d,
		AdminValidator: adminV,
		RepoValidator:  repoV,
		Storage:        storage,
		Producer:       nsqpb.GetInitProducer(),
		hhBaseUrl:      hhBaseUrl,
	}

	anyCredAPI := anycred.AnyCredAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	buildAPI := BuildAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		Deserializer:   d,
		Producer:       nsqpb.GetInitProducer(),
		OcyValidator:   buildconfigvalidator.GetOcelotValidator(),
	}

	appleDevSecretAPI := AppleDevSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	artifactRepoSecretAPI := ArtifactRepoSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		RepoValidator:  repoV,
	}

	genericSecretAPI := GenericSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	kubernetesSecretAPI := KubernetesSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	notifierSecretAPI := NotifierSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	pollScheduleAPI := PollScheduleAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		Producer:       nsqpb.GetInitProducer(),
	}

	repoInterfaceAPI := RepoInterfaceAPI {
		RemoteConfig:   config,
		Storage:        storage,	
		HhBaseUrl:      hhBaseUrl,
	}

	sshSecretAPI := SshSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	statusInterfaceAPI := StatusInterfaceAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	vcsSecretAPI := VcsSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		AdminValidator: adminV,
	}

	return &OcelotServerAPI{ 
		guideOcelotServer,
		anyCredAPI,
		buildAPI,
		appleDevSecretAPI,
		artifactRepoSecretAPI,
		genericSecretAPI,
		kubernetesSecretAPI,
		notifierSecretAPI,
		pollScheduleAPI,
		repoInterfaceAPI,
		sshSecretAPI,
		statusInterfaceAPI,
		vcsSecretAPI,
	}
}
