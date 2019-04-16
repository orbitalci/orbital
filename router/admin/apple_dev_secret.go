package admin

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"github.com/shankj3/go-til/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/ocelot/build/helpers/ioshelper"
)

type AppleDevSecret interface {
	SetAppleCreds(context.Context, *pb.AppleCreds) (*empty.Empty, error)
	GetAppleCreds(context.Context, *empty.Empty) (*pb.AppleCredsWrapper, error)
	GetAppleCred(context.Context, *pb.AppleCreds) (*pb.AppleCreds, error)
	UpdateAppleCreds(context.Context, *pb.AppleCreds) (*empty.Empty, error)
	AppleCredExists(context.Context, *pb.AppleCreds) (*pb.Exists, error)
}

func appleNastiness(zipFile []byte, devProfilePassword string) (parsed []byte, err error) {
	appleKeychain, err := ioshelper.UnpackAppleDevAccount(zipFile, devProfilePassword)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't deal with this zip file...")
		return nil, status.Error(codes.InvalidArgument, "could not unpack developeraccount zip to keychain, error is :"+err.Error())
	}
	return appleKeychain, nil
}

func (g *guideOcelotServer) SetAppleCreds(ctx context.Context, creds *pb.AppleCreds) (*empty.Empty, error) {
	vempty := &empty.Empty{}
	if creds.GetSubType().Parent() != pb.CredType_APPLE {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of apple type: "+strings.Join(pb.CredType_APPLE.SubtypesString(), " | "))
	}
	var err error
	creds.AppleSecrets, err = appleNastiness(creds.AppleSecrets, creds.AppleSecretsPassword)
	if err != nil {
		return nil, err
	}

	if err := SetupRCCCredentials(g.RemoteConfig, g.Storage, creds); err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return vempty, status.Error(codes.InvalidArgument, "Apple creds upload failed validation, errors are: "+err.Error())
		}
		return vempty, status.Error(codes.Internal, "Apple creds could not be uploaded, error is: "+err.Error())
	}
	log.Log().Info("unpacked & stored apple dev profile")
	return vempty, nil
}

func (g *guideOcelotServer) GetAppleCreds(ctx context.Context, empty2 *empty.Empty) (*pb.AppleCredsWrapper, error) {
	wrapper := &pb.AppleCredsWrapper{}
	credz, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_APPLE, true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return wrapper, status.Errorf(codes.Internal, "unable to get apple creds! error: %s", err.Error())
	}
	for _, v := range credz {
		wrapper.AppleCreds = append(wrapper.AppleCreds, v.(*pb.AppleCreds))
	}
	if len(wrapper.AppleCreds) == 0 {
		return nil, status.Error(codes.NotFound, "no apple creds found")
	}
	return wrapper, nil
}

func (g *guideOcelotServer) GetAppleCred(ctx context.Context, creds *pb.AppleCreds) (*pb.AppleCreds, error) {
	creddy, err := g.getAnyCred(creds)
	if err != nil {
		return nil, err
	}
	apple, ok := creddy.(*pb.AppleCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to cast as apple creds")
	}
	return apple, nil
}

func (g *guideOcelotServer) UpdateAppleCreds(ctx context.Context, creds *pb.AppleCreds) (*empty.Empty, error) {
	var err error
	creds.AppleSecrets, err = appleNastiness(creds.AppleSecrets, creds.AppleSecretsPassword)
	if err != nil {
		return nil, err
	}
	return g.updateAnyCred(ctx, creds)
}

func (g *guideOcelotServer) AppleCredExists(ctx context.Context, creds *pb.AppleCreds) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}
