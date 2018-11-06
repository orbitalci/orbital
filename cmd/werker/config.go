package main

import (
	"strings"

	"github.com/namsral/flag"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/storage"
	"github.com/shankj3/ocelot/version"

	"errors"
	"os"
)

const (
	defaultServicePort = "9090"
	defaultGrpcPort    = "9099"
	defaultWerkerType  = "docker"
	defaultStorage     = "filesystem"
)

func strToWerkType(str string) models.WerkType {
	switch str {
	case "k8s", "kubernetes":
		return models.Kubernetes
	case "docker":
		return models.Docker
	case "ssh":
		return models.SSH
	case "exec":
		return models.Exec
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
	//ServicePort string
	//GrpcPort    string
	//WerkerType  models.WerkType
	//WerkerUuid		uuid.UUID
	*models.WerkerFacts
	WerkerName string
	// list of tags for this build node
	tags []string
	//werkerProcessor builder.Processor
	LogLevel string
	//LoopBackIp      string
	RemoteConfig cred.CVRemoteConfig
}

// GetConf sets the configuration for the Werker. Its not thread safe, but that's
// alright because it only happens on startup of the application
func GetConf() (*WerkerConf, error) {
	werker := &WerkerConf{WerkerFacts: models.NewFacts()}
	werkerName, _ := os.Hostname()
	var werkerTypeStr string
	var consuladdr string
	var consulport int
	var tags string

	flrg := flag.NewFlagSet("werker", flag.ExitOnError)
	flrg.StringVar(&werkerTypeStr, "type", defaultWerkerType, "type of werker, kubernetes|docker|ssh")
	flrg.StringVar(&werker.WerkerName, "name", werkerName, "if wish to identify as other than hostname")
	flrg.StringVar(&werker.ServicePort, "ws-port", defaultServicePort, "port to run websocket service on. default 9090")
	flrg.StringVar(&werker.GrpcPort, "grpc-port", defaultGrpcPort, "port to run grpc server on. default 9099")
	flrg.StringVar(&werker.LogLevel, "log-level", "info", "log level")
	flrg.BoolVar(&werker.Dev, "dev", false, "run dev mode")
	flrg.StringVar(&werker.RegisterIP, "register-ip", "localhost", "ip to register with consul when picking up builds")
	flrg.StringVar(&werker.LoopbackIp, "loopback-ip", "172.17.0.1", "ip to use for spawned containers to successfully contact the host. "+
		"This may be different for different container systems / host machines. For example, when using docker for mac the loopback-ip would be docker.for.mac.localhost")
	flrg.StringVar(&consuladdr, "consul-host", "localhost", "address of consul")
	flrg.IntVar(&consulport, "consul-port", 8500, "port of consul")
	flrg.StringVar(&tags, "tags", "", "comma separated list of tags for this build node")
	// set flags for disk utility checks
	flrg.StringVar(&werker.DiskUtilityHealthCheck.PauseThreshold, "disk-pause-threshold", "1G", "How much free disk can be left before the werker will error out and stop consuming messasges")
	flrg.StringVar(&werker.DiskUtilityHealthCheck.Path, "disk-pause-path", "","Path at which to check for disk-pause-threshold")
	// set flag for interval on which to run healthcheck
	flrg.IntVar(&werker.HealthCheckInterval, "healthcheck-interval", 60, "Interval on which to make sure that remote config, storage system, and disk is not full")
	// ssh werker configuration
	werker.Ssh.SetFlags(flrg)
	flrg.Parse(os.Args[1:])
	version.MaybePrintVersion(flrg.Args())
	werker.tags = strings.Split(tags, ",")
	werker.WerkerType = strToWerkType(werkerTypeStr)
	if werker.WerkerType == -1 {
		return nil, errors.New("werker type can only be: k8s, kubernetes, docker, ssh, exec")
	}
	if werker.WerkerType == models.SSH && !werker.Ssh.IsValid() {
		return nil, errors.New("if werker type is ssh, then -ssh-port, -ssh-host, -ssh-user, and either -ssh-private-key or -ssh-password are required fields")
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
