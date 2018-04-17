package integrations

import (
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

type StringIntegrator interface {
	GenerateIntegrationString([]pb.OcyCredder) (string, error)
	String() string
	SubType() pb.SubCredType
	//GetThemCreds(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) ([]pb.OcyCredder, error)
}
