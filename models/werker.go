package models

import (
	"context"

	"github.com/google/uuid"
)

type WerkType int

const (
	Kubernetes WerkType = iota
	Docker
	SSH
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

// GetOcyPrefixFromWerkerType will return "" for anything that runs in a container because root access can be assumed
// If it is running with the SSH connection (ie mac builds) then it will find the home direc and use that as the prefix for the .ocelot directory
func GetOcyPrefixFromWerkerType(wt WerkType) string {
	switch wt {
	case Docker:
		return ""
	case Kubernetes:
		return ""
	case SSH:
		return "/tmp"
	default:
		return ""
	}
}

func NewFacts() *WerkerFacts {
	return &WerkerFacts{Ssh:&SSHFacts{}}
}

// WerkerFacts is a struct for the configurations in werker that affect actual builds.
// Think of it like gather facts w/ ansible.
type WerkerFacts struct {
	Uuid 	       uuid.UUID
	WerkerType     WerkType
	LoopbackIp     string
	RegisterIP     string
	ServicePort    string
	GrpcPort       string
	// set dev mode
	Dev			   bool
	// this is only for SSH type werkers
	Ssh            *SSHFacts
}

// When a werker starts up as an SSH werker, it will also need to be initialized with these fields so it knows
//   what to connect to
type SSHFacts struct {
	User      string
	Host      string
	Port      int
	KeyFP     string
	Password  string
}

func (sf *SSHFacts) IsValid() bool {
	if sf.User == "" || sf.Host == "" || sf.Port == 0 {
		return false
	}
	if sf.Password == "" && sf.KeyFP == "" {
		return false
	}
	return true
}