package integrations

import (
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

// Interface for injecting data during the build process, used in runIntegrations in build/launcher/makeitso.go
type StringIntegrator interface {
	//GenerateIntegrationString is executed fourth, and provides the input for MakeBashable
	GenerateIntegrationString([]pb.OcyCredder) (string, error)
	//String is executed second
	String() string
	//SubType is executed third
	SubType() pb.SubCredType
	//MakeBashable is run after GenerateIntegrationString, along with GetEnv
	MakeBashable(input string) []string
	//GetEnv is run after GenerateIntegrationString, along with MakeBashable
	GetEnv() []string
	//IsRelevant executed first
	IsRelevant(wc *pb.BuildConfig) bool
}
