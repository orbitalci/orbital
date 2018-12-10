package integrations

import (
	"github.com/level11consulting/ocelot/models/pb"
)

type integrator interface {
	//IsRelevant is for determining if this integrator should execute
	IsRelevant(wc *pb.BuildConfig) bool
	//String representation of integrator
	String() string
}

// Interface for injecting data during the build process, used in runIntegrations in build/launcher/makeitso.go
type StringIntegrator interface {
	//GenerateIntegrationString is executed fourth, and provides the input for MakeBashable
	GenerateIntegrationString([]pb.OcyCredder) (string, error)
	//String is executed second
	//String() string
	//SubType is executed third
	SubType() pb.SubCredType
	//MakeBashable is run after GenerateIntegrationString, along with GetEnv
	MakeBashable(input string) []string
	//GetEnv is run after GenerateIntegrationString, along with MakeBashable
	GetEnv() []string
	////IsRelevant executed first
	//IsRelevant(wc *pb.BuildConfig) bool
	integrator
}

// BinaryIntegrator is an interface for determining when/how to download binaries
type BinaryIntegrator interface {
	//GenerateDownloadBashables should create the download binary command
	GenerateDownloadBashables() []string
	integrator
}
