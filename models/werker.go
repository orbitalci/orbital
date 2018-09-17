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
	Exec
)

func (w WerkType) String() string {
	switch w {
	case Kubernetes:
		return "Kubernetes"
	case Docker:
		return "Docker"
	case SSH:
		return "SSH"
	case Exec:
		return "Exec"
	default:
		panic("inconceivable!!")
	}
}

// Transport struct is for the Transport channel that will interact with the streaming side of the service
// to stream results back to the admin. It sends just enough to be unique, the hash that triggered the build
// and the InfoChan which the builder will write to.
type Transport struct {
	Hash     string
	InfoChan chan []byte
	DbId     int64
}

type BuildContext struct {
	Hash       string
	Context    context.Context
	CancelFunc func()
}

func NewFacts() *WerkerFacts {
	return &WerkerFacts{Ssh: &SSHFacts{}}
}

// WerkerFacts is a struct for the configurations in werker that affect actual builds.
// Think of it like gather facts w/ ansible.
type WerkerFacts struct {
	Uuid        uuid.UUID
	WerkerType  WerkType
	LoopbackIp  string
	RegisterIP  string
	ServicePort string
	GrpcPort    string
	// set dev mode
	Dev bool
	// this is only for SSH type werkers
	Ssh *SSHFacts
}

// When a werker starts up as an SSH werker, it will also need to be initialized with these fields so it knows
//   what to connect to
type SSHFacts struct {
	User     string
	Host     string
	Port     int
	KeyFP    string
	Password string
	// KeepRepo; if true then the repositories will be left on machine and new commits will be checked out instead of re-cloned? idk. maybe not.
	KeepRepo bool
}

func (sf *SSHFacts) SetFlags(flg Flagger) {
	flg.IntVar(&sf.Port, "ssh-port", 22, "port to ssh to for build exectuion | ONLY VALID FOR SSH TYPE WERKERS")
	flg.StringVar(&sf.Host, "ssh-host", "", "host to ssh to for build execution | ONLY VALID FOR SSH TYPE WERKERS")
	flg.StringVar(&sf.KeyFP, "ssh-private-key", "", "private key for using ssh for build execution | ONLY VALID FOR SSH TYPE WERKERS")
	flg.StringVar(&sf.User, "ssh-user", "root", "ssh user for build execution | ONLY VALID FOR SSH TYPE WERKERS")
	flg.StringVar(&sf.Password, "ssh-password", "", "password for ssh user if no key file | ONLY VALID FOR SSH TYPE WERKERS")
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
