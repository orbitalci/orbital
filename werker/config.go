package main

import (
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
	defaultWerkerType  = "docker"
)

func StrToWerkType(str string) WerkType {
	switch str {
	case "k8s", "kubernetes": return Kubernetes
	case "docker": 			  return Docker
	default: 				  return -1
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
	flag.StringVar(&werkerTypeStr, "werker_type", defaultWerkerType, "type of werker, kubernetes or docker")
	flag.StringVar(&werker.werkerName,"werker_name", werkerName, "if wish to identify as other than hostname")
	flag.StringVar(&werker.servicePort, "werker_port", defaultServicePort, "port to run service on. default 9090")
	flag.StringVar(&werker.logLevel, "log-level", "info", "log level")
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