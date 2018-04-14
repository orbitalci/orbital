package models

import (
	"context"

	"github.com/google/uuid"
)

type WerkType int

const (
	Kubernetes WerkType = iota
	Docker
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

// WerkerFacts is a struct for the configurations in werker that affect actual builds.
// Think of it like gather facts w/ ansible.
type WerkerFacts struct {
	Uuid 	    	uuid.UUID
	WerkerType           WerkType
	LoopbackIp  	string
	ServicePort    string
	GrpcPort       string
}