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
	*guideOcelotServer
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



	return &OcelotServerAPI{ guideOcelotServer }
}
