package admin

import (
	"context"

	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/models"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shankj3/ocelot/build"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

//this is our grpc server, it responds to client requests
type guideOcelotServer struct {
	RemoteConfig   cred.CVRemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *cred.AdminValidator
	RepoValidator  *cred.RepoValidator
	OcyValidator   *build.OcelotValidator
	Storage        storage.OcelotStorage
	Producer       nsqpb.Producer
	handler        models.VCSHandler
}

// for checking if the server is reachable
func (g *guideOcelotServer) CheckConn(ctx context.Context, msg *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func NewGuideOcelotServer(config cred.CVRemoteConfig, d *deserialize.Deserializer, adminV *cred.AdminValidator, repoV *cred.RepoValidator, storage storage.OcelotStorage) pb.GuideOcelotServer {
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
	}
	return guideOcelotServer
}
