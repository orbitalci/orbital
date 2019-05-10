package anycred

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/orbitalci/models/pb"
	"github.com/level11consulting/orbitalci/storage"
	"github.com/level11consulting/orbitalci/server/config"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnyCred interface {
	CheckAnyCredExists(ctx context.Context, creds pb.OcyCredder) (*pb.Exists, error)
	UpdateAnyCred(ctx context.Context, creds pb.OcyCredder) (*empty.Empty, error)
}

type AnyCredAPI struct {
	AnyCred
	RemoteConfig   config.CVRemoteConfig
	Storage        storage.OcelotStorage
}

func (g *AnyCredAPI) CheckAnyCredExists(ctx context.Context, creds pb.OcyCredder) (*pb.Exists, error) {
	exists, err := g.Storage.CredExists(creds)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Unable to reach cred table to check if cred %s/%s/%s exists. Error: %s", creds.GetAcctName(), creds.GetSubType().String(), creds.GetIdentifier(), err.Error())
	}
	return &pb.Exists{Exists: exists}, nil
}

func (g *AnyCredAPI) UpdateAnyCred(ctx context.Context, creds pb.OcyCredder) (*empty.Empty, error) {
	if err := g.RemoteConfig.UpdateCreds(g.Storage, creds); err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "%s cred failed validation. Errors are: %s", creds.GetSubType().Parent(), err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Unavailable, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *AnyCredAPI) DeleteAnyCred(ctx context.Context, creds pb.OcyCredder, parentType pb.CredType) (*empty.Empty, error) {
	// make sure we have all the fields we need to be able to accurately delete the credential.
	// try to intelligently deduce what subType teh cred is, but error out if that isn't possible
	empti := &empty.Empty{}
	var errmsg string
	if creds.GetIdentifier() == "" || creds.GetAcctName() == "" {
		errmsg += "identifier and acctName are required fields; "
	}
	if creds.GetSubType() == pb.SubCredType_NIL_SCT {
		if len(parentType.Subtypes()) == 1 {
			creds.SetSubType(parentType.Subtypes()[0])
		} else {
			errmsg += "subType must be set since there is more than one sub type to this parent type " + parentType.String() + ": " + strings.Join(parentType.SubtypesString(), "|")
		}
	}
	if errmsg != "" {
		return empti, status.Error(codes.InvalidArgument, errmsg)
	}

	if exists, _ := g.Storage.CredExists(creds); !exists {
		return empti, status.Error(codes.NotFound, "not found")
	}

	if err := g.RemoteConfig.DeleteCred(g.Storage, creds); err != nil {
		return empti, status.Error(codes.Internal, err.Error())
	}
	return empti, nil

}
