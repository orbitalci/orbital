package credentials

import (
	"strings"

	"github.com/level11consulting/ocelot/models/pb"

	"github.com/pkg/errors"
)

//validator for all admin related stuff
type AdminValidator struct{}

func GetValidator() *AdminValidator {
	return &AdminValidator{}
}

//validates config and returns json formatted error
func (adminValidator AdminValidator) ValidateConfig(adminCreds *pb.VCSCreds) error {
	if len(adminCreds.AcctName) == 0 {
		return errors.New("acctName is required")
	}
	if len(adminCreds.ClientId) == 0 {
		return errors.New("clientId is required")
	}
	if len(adminCreds.ClientSecret) == 0 {
		return errors.New("clientSecret is required")
	}
	if len(adminCreds.TokenURL) == 0 {
		return errors.New("tokenURL is required")
	}
	switch adminCreds.SubType {
	case pb.SubCredType_NIL_SCT:
		return errors.New("SUB CRED TYPE WAS NOT INSTANTIATED PROPERLY")
	case pb.SubCredType_BITBUCKET:
		return nil
	case pb.SubCredType_GITHUB:
		return nil
	default:
		return errors.Errorf("creds must be one of the following type: %s", strings.Join(pb.CredType_VCS.SubtypesString(), "|"))
	}
	return nil
}

type RepoValidator struct{}

func GetRepoValidator() *RepoValidator {
	return &RepoValidator{}
}

// RepoValidator.ValidateConfig validates config and returns an error if it does not meet spec
func (RepoValidator) ValidateConfig(repoCreds *pb.RepoCreds) error {
	if len(repoCreds.Password) == 0 {
		return errors.New("password is required")
	}
	if len(repoCreds.RepoUrl) == 0 {
		return errors.New("field repoUrl is required")
	}
	if len(repoCreds.AcctName) == 0 {
		return errors.New("field acctName is required")
	}
	if len(repoCreds.Username) == 0 {
		return errors.New("field username is required")
	}
	switch {
	case pb.Contains(repoCreds.SubType, pb.CredType_REPO.Subtypes()):
		return nil
	default:
		return errors.New("repo creds must be one of the following type: nexus | maven | docker")
	}
}
