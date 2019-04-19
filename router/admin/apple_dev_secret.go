package admin

import (
	"fmt"
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"github.com/shankj3/go-til/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/ocelot/build/helpers/ioshelper"
	"github.com/level11consulting/ocelot/secret/legacy"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/router/admin/anycred"
)

type AppleDevSecret interface {
	SetAppleCreds(context.Context, *pb.AppleCreds) (*empty.Empty, error)
	GetAppleCreds(context.Context, *empty.Empty) (*pb.AppleCredsWrapper, error)
	GetAppleCred(context.Context, *pb.AppleCreds) (*pb.AppleCreds, error)
	UpdateAppleCreds(context.Context, *pb.AppleCreds) (*empty.Empty, error)
	AppleCredExists(context.Context, *pb.AppleCreds) (*pb.Exists, error)
}

type AppleDevSecretAPI struct {
	AppleDevSecret
	RemoteConfig   config.CVRemoteConfig
	Storage        storage.OcelotStorage
}

func appleNastiness(zipFile []byte, devProfilePassword string) (parsed []byte, err error) {
	appleKeychain, err := ioshelper.UnpackAppleDevAccount(zipFile, devProfilePassword)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't deal with this zip file...")
		return nil, status.Error(codes.InvalidArgument, "could not unpack developeraccount zip to keychain, error is :"+err.Error())
	}
	return appleKeychain, nil
}

func (g *AppleDevSecretAPI) SetAppleCreds(ctx context.Context, creds *pb.AppleCreds) (*empty.Empty, error) {
	vempty := &empty.Empty{}
	if creds.GetSubType().Parent() != pb.CredType_APPLE {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of apple type: "+strings.Join(pb.CredType_APPLE.SubtypesString(), " | "))
	}
	var err error
	creds.AppleSecrets, err = appleNastiness(creds.AppleSecrets, creds.AppleSecretsPassword)
	if err != nil {
		return nil, err
	}

	if err := legacy.SetupRCCCredentials(g.RemoteConfig, g.Storage, creds); err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return vempty, status.Error(codes.InvalidArgument, "Apple creds upload failed validation, errors are: "+err.Error())
		}
		return vempty, status.Error(codes.Internal, "Apple creds could not be uploaded, error is: "+err.Error())
	}
	log.Log().Info("unpacked & stored apple dev profile")
	return vempty, nil
}

func (g *AppleDevSecretAPI) GetAppleCreds(ctx context.Context, empty2 *empty.Empty) (*pb.AppleCredsWrapper, error) {
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

func (g *AppleDevSecretAPI) GetAppleCred(ctx context.Context, creds *pb.AppleCreds) (*pb.AppleCreds, error) {
	creddy, err := g.GetAnyCred(creds)
	if err != nil {
		return nil, err
	}
	apple, ok := creddy.(*pb.AppleCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to cast as apple creds")
	}
	return apple, nil
}

func (g *AppleDevSecretAPI) UpdateAppleCreds(ctx context.Context, creds *pb.AppleCreds) (*empty.Empty, error) {
	var err error
	creds.AppleSecrets, err = appleNastiness(creds.AppleSecrets, creds.AppleSecretsPassword)
	if err != nil {
		return nil, err
	}

	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.UpdateAnyCred(ctx, creds)
}

func (g *AppleDevSecretAPI) AppleCredExists(ctx context.Context, creds *pb.AppleCreds) (*pb.Exists, error) {
	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.CheckAnyCredExists(ctx, creds)
}

func (g *AppleDevSecretAPI) GetAnyCred(credder pb.OcyCredder) (pb.OcyCredder, error) {
	if credder.GetSubType() == 0 || credder.GetAcctName() == "" || credder.GetIdentifier() == "" {
		return nil, status.Error(codes.InvalidArgument, "subType, acctName, and identifier are required fields")
	}
	creddy, err := g.RemoteConfig.GetCred(g.Storage, credder.GetSubType(), credder.GetIdentifier(), credder.GetAcctName(), true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Credential %s/%s of Type %s Not Found", credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType()))
		}
		if _, ok := err.(*pb.ValidationErr); ok {
			return nil, status.Error(codes.InvalidArgument, "Invalid arguments, error: "+err.Error())
		}
		return nil, status.Error(codes.Unavailable, "Credential interface not available, error: "+err.Error())
	}
	return creddy, nil
}