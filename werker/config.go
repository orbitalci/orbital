package werker

import (
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"errors"
	"github.com/namsral/flag"
	"os"
)

type WerkType int

const (
	Kubernetes WerkType = iota
	Docker
)

const (
	defaultServicePort = "9090"
	defaultGrpcPort    = "9099"
	defaultWerkerType  = "docker"
	defaultStorage     = "filesystem"
)

func strToWerkType(str string) WerkType {
	switch str {
	case "k8s", "kubernetes":
		return Kubernetes
	case "docker":
		return Docker
	default:
		return -1
	}
}

func strToStorageImplement(str string) storage.BuildOut {
	switch str {
	case "filesystem":
		return storage.NewFileBuildStorage("")
	// as more are written, include here
	default:
		return storage.NewFileBuildStorage("")
	}
}

// WerkerConf is all the configuration for the Werker to do its job properly. this is where the
// storage type is set (ie filesystem, etc..) and the processor is set (ie Docker, kubernetes, etc..)
type WerkerConf struct {
	ServicePort     string
	grpcPort        string
	WerkerName      string
	werkerType      WerkType
	//werkerProcessor builder.Processor
	LogLevel        string
	RegisterIP     string
	RemoteConfig   cred.CVRemoteConfig
}

// GetConf sets the configuration for the Werker. Its not thread safe, but that's
// alright because it only happens on startup of the application
func GetConf() (*WerkerConf, error) {
	werker := &WerkerConf{}
	werkerName, _ := os.Hostname()
	var werkerTypeStr string
	var storageTypeStr string
	var consuladdr string
	var consulport int
	//todo: idk about this env prefix thing, might not be necessary
	flrg := flag.NewFlagSet("werker", flag.ExitOnError)
	flrg.StringVar(&werkerTypeStr, "type", defaultWerkerType, "type of werker, kubernetes or docker")
	flrg.StringVar(&werker.WerkerName, "name", werkerName, "if wish to identify as other than hostname")
	flrg.StringVar(&werker.ServicePort, "ws-port", defaultServicePort, "port to run websocket service on. default 9090")
	flrg.StringVar(&werker.grpcPort, "grpc-port", defaultGrpcPort, "port to run grpc server on. default 9099")
	flrg.StringVar(&werker.LogLevel, "log-level", "info", "log level")
	flrg.StringVar(&storageTypeStr, "storage-type", defaultStorage, "storage type to use for build info, available: [filesystem")
	flrg.StringVar(&werker.RegisterIP, "register-ip", "localhost", "ip to register with consul when picking up builds")
	flrg.StringVar(&consuladdr, "consul-host", "localhost", "address of consul")
	flrg.IntVar(&consulport, "consul-port", 8500, "port of consul")
	flrg.Parse(os.Args[1:])
	werker.werkerType = strToWerkType(werkerTypeStr)
	if werker.werkerType == -1 {
		return nil, errors.New("werker type can only be: k8s, kubernetes, docker")
	}
	if werker.WerkerName == "" {
		return nil, errors.New("could not get hostname from os.hostname() and no werker_name given")
	}
	rc, err := cred.GetInstance(consuladdr, consulport, "")
	if err != nil {
		return nil, errors.New("could not get instance of remote config; err: " + err.Error())
	}

	werker.RemoteConfig = rc
	return werker, nil
}
