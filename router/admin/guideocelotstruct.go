package admin

import (
	"context"

	"github.com/level11consulting/ocelot/models"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/build"
	creds "github.com/level11consulting/ocelot/common/credentials"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/storage"
)

//this is our grpc server, it responds to client requests
type guideOcelotServer struct {
	RemoteConfig   config.CVRemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *creds.AdminValidator
	RepoValidator  *creds.RepoValidator
	OcyValidator   *build.OcelotValidator
	Storage        storage.OcelotStorage
	Producer       nsqpb.Producer
	handler        models.VCSHandler
	hhBaseUrl      string
}

// for checking if the server is reachable
func (g *guideOcelotServer) CheckConn(ctx context.Context, msg *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func NewGuideOcelotServer(config config.CVRemoteConfig, d *deserialize.Deserializer, adminV *creds.AdminValidator, repoV *creds.RepoValidator, storage storage.OcelotStorage, hhBaseUrl string) pb.GuideOcelotServer {
	// changing to this style of instantiation cuz thread safe (idk read it on some best practices, it just looks
	// purdier to me anyway
	guideOcelotServer := &guideOcelotServer{
		OcyValidator:   build.GetOcelotValidator(),
		RemoteConfig:   config,
		Deserializer:   d,
		AdminValidator: adminV,
		RepoValidator:  repoV,
		Storage:        storage,
		Producer:       nsqpb.GetInitProducer(),
		hhBaseUrl:      hhBaseUrl,
	}
	return guideOcelotServer
}
