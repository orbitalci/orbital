package credentials

import (
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"

	"github.com/pkg/errors"
)

// GetVcsCreds will retrieve a VCSCred for account name / bitbucket vcs type
func GetVcsCreds(store storage.CredTable, repoFullName string, remoteConfig CVRemoteConfig, credType pb.SubCredType) (*pb.VCSCreds, error) {
	acctName, _, err := common.GetAcctRepo(repoFullName)
	if err != nil {
		return nil, err
	}
	if credType == pb.SubCredType_NIL_SCT {
		credType, err = store.GetVCSTypeFromAccount(acctName)
		if err != nil {
			// don't wrap error here because we want to do type checking of the error
			// if the caller of this function will be client facing
			return nil, err
		}
	}
	identifier, err := pb.CreateVCSIdentifier(credType, acctName)
	if err != nil {
		return nil, err
	}
	bbCreds, err := remoteConfig.GetCred(store, credType, identifier, acctName, false)
	if err != nil {
		return nil, err
	}
	vcs, ok := bbCreds.(*pb.VCSCreds)
	if !ok {
		return nil, errors.New("could not cast as vcs creds")
	}
	return vcs, err
}
