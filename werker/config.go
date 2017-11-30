package main

import (
	"errors"
	"github.com/namsral/flag"
	"github.com/shankj3/ocelot/util/storage"
	"github.com/shankj3/ocelot/werker/processors"
	"os"
)

type WerkType int

const (
	Kubernetes WerkType = iota
	Docker
)

const (
	defaultServicePort = "9090"
	defaultWerkerType  = "docker"
	defaultStorage     = "filesystem"
)

func strToWerkType(str string) WerkType {
	switch str {
	case "k8s", "kubernetes": return Kubernetes
	case "docker": 			  return Docker
	default: 				  return -1
	}
}

func strToStorageImplement(str string) storage.BuildOutputStorage {
	switch str {
	case "filesystem": return storage.NewFileBuildStorage("")
	// as more are written, include here
	default: 		   return storage.NewFileBuildStorage("")
	}
}

// WerkerConf is all the configuration for the Werker to do its job properly. this is where the
// storage type is set (ie filesystem, etc..) and the processor is set (ie Docker, kubernetes, etc..)
type WerkerConf struct {
	servicePort   	string
	werkerName    	string
	werkerType    	WerkType
	werkerProcessor processors.Processor
	storage 		storage.BuildOutputStorage
	logLevel 	  	string
}

// GetConf sets the configuration for the Werker. Its not thread safe, but that's
// alright because it only happens on startup of the application
func GetConf() (*WerkerConf, error) {
	werker := &WerkerConf{}
	werkerName, _ := os.Hostname()
	var werkerTypeStr string
	var storageTypeStr string
	flag.StringVar(&werkerTypeStr, "werker_type", defaultWerkerType, "type of werker, kubernetes or docker")
	flag.StringVar(&werker.werkerName,"werker_name", werkerName, "if wish to identify as other than hostname")
	flag.StringVar(&werker.servicePort, "werker_port", defaultServicePort, "port to run service on. default 9090")
	flag.StringVar(&werker.logLevel, "log-level", "info", "log level")
	flag.StringVar(&storageTypeStr, "storage-type", defaultStorage, "storage type to use for build info, available: [filesystem")
	flag.Parse()
	werker.werkerType = strToWerkType(werkerTypeStr)
	if werker.werkerType == -1 {
		return nil, errors.New("werker type can only be: k8s, kubernetes, docker")
	}
	if werker.werkerName == "" {
		return nil, errors.New("could not get hostname from os.hostname() and no werker_name given")
	}
	werker.storage = strToStorageImplement(storageTypeStr)

	switch werker.werkerType {
	case Kubernetes:
		werker.werkerProcessor = &processors.K8Proc{}
	case Docker:
		werker.werkerProcessor = &processors.DockProc{}
	}

	return werker, nil
}