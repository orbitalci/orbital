package main

import (
	"errors"
	"github.com/namsral/flag"
	"github.com/shankj3/ocelot/util/storage"
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

func StrToWerkType(str string) WerkType {
	switch str {
	case "k8s", "kubernetes": return Kubernetes
	case "docker": 			  return Docker
	default: 				  return -1
	}
}

func StrToStorageImplement(str string) storage.BuildOutputStorage {
	switch str {
	case "filesystem": return &storage.FileBuildStorage{}
	// as more are written, include here
	default: 		   return &storage.FileBuildStorage{}
	}
}

type WerkerConf struct {
	servicePort   string
	werkerName    string
	werkerType    WerkType
	logLevel 	  string
}

func GetConf() (*WerkerConf, error) {
	werker := &WerkerConf{}
	werkerName, _ := os.Hostname()
	var werkerTypeStr string
	var storageTypeStr string
	flag.StringVar(&werkerTypeStr, "werker_type", defaultWerkerType, "type of werker, kubernetes or docker")
	flag.StringVar(&werker.werkerName,"werker_name", werkerName, "if wish to identify as other than hostname")
	flag.StringVar(&werker.servicePort, "werker_port", defaultServicePort, "port to run service on. default 9090")
	flag.StringVar(&werker.logLevel, "log-level", "info", "log level")
	flag.StringVar(&storageTypeStr, "storage-type", defaultStorage, "storage type to use for build info, availabe: [filesystem")
	flag.Parse()
	werker.werkerType = StrToWerkType(werkerTypeStr)
	if werker.werkerType == -1 {
		return nil, errors.New("werker type can only be: k8s, kubernetes, docker")
	}
	if werker.werkerName == "" {
		return nil, errors.New("could not get hostname from os.hostname() and no werker_name given")
	}
	return werker, nil
}