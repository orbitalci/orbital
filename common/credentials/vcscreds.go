package credentials

import (
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"

	"errors"
)

// GetVcsCreds will retrieve a VCSCred for account name / bitbucket vcs type
func GetVcsCreds(store storage.CredTable, repoFullName string, remoteConfig CVRemoteConfig) (*pb.VCSCreds, error) {
	acctName, _, err := common.GetAcctRepo(repoFullName)
	if err != nil {
		return nil, err
	}
	identifier, err := pb.CreateVCSIdentifier(pb.SubCredType_BITBUCKET, acctName)
	if err != nil {
		return nil, err
	}
	bbCreds, err := remoteConfig.GetCred(store, pb.SubCredType_BITBUCKET, identifier, acctName, false)
	if err != nil {
		return nil, err
	}
	vcs, ok := bbCreds.(*pb.VCSCreds)
	if !ok {
		return nil, errors.New("could not cast as vcs creds")
	}
	return vcs, err
}
