package integrations

import (
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

// this is what you have to implement to be able to be run in the makeitso function doIntegrations
type StringIntegrator interface {
	GenerateIntegrationString([]pb.OcyCredder) (string, error)
	String() string
	SubType() pb.SubCredType
	MakeBashable(input string) []string
	GetEnv() []string
	IsRelevant(wc *pb.BuildConfig) bool
}
