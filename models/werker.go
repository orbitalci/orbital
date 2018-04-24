package models

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
)

type WerkType int

const (
	Kubernetes WerkType = iota
	Docker
	Host
)


// Transport struct is for the Transport channel that will interact with the streaming side of the service
// to stream results back to the admin. It sends just enough to be unique, the hash that triggered the build
// and the InfoChan which the builder will write to.
type Transport struct {
	Hash     string
	InfoChan chan []byte
	DbId     int64
}

type BuildContext struct {
	Hash string
	Context context.Context
	CancelFunc func()
}

func GetOcyPrefixFromWerkerType(wt WerkType) string {
	switch wt {
	case Docker:
		return ""
	case Kubernetes:
		return ""
	case Host:
		dir, err := homedir.Dir()
		if err != nil {
			panic("Couldn't get home directory! " + err.Error())
		}
		return dir
	default:
		return ""
	}
}

// WerkerFacts is a struct for the configurations in werker that affect actual builds.
// Think of it like gather facts w/ ansible.
type WerkerFacts struct {
	Uuid 	    	uuid.UUID
	WerkerType           WerkType
	LoopbackIp  	string
	ServicePort    string
	GrpcPort       string
}