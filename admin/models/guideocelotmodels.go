package models

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"fmt"
	"google.golang.org/grpc"
	"strings"
)

func NewRepoCreds() *RepoCreds {
	return &RepoCreds{
		RepoUrl: make(map[string]string),
	}
}

// these methods are attached to the proto object RepoCreds
func (m *RepoCreds) SetAcctNameAndType(name string, typ string) {
	m.AcctName = name
	m.Type = typ
}

func (m *RepoCreds) BuildCredPath(credType string, acctName string) string {
	return fmt.Sprintf("%s/repo/%s/%s", "creds", acctName, credType)
}

func (m *RepoCreds) SetSecret(secret string) {
	m.Password = secret
}

func (m *RepoCreds) GetClientSecret() string {
	return m.Password
}

func (m *RepoCreds) SetAdditionalFields(infoType string, val string) {
	if strings.Contains(infoType, "repourl") {
		paths := strings.Split(infoType, "/")
		if len(paths) > 2 {
			panic("WHAT THE FUCK?")
		}
		m.RepoUrl[paths[1]] = val
	}
	if infoType == "username" {
		m.Username = val
	}
}

func (m *RepoCreds) AddAdditionalFields(consule *consul.Consulet, path string) (err error) {
	if err := consule.AddKeyValue(path + "/username", []byte(m.Username)); err != nil {
		return err
	}
	for reponame, url := range m.RepoUrl {
		if err = consule.AddKeyValue(path + "/repourl/" + reponame, []byte(url)); err != nil {
			return err
		}
	}
	return err
}

func (m *RepoCreds) Spawn() cred.RemoteConfigCred {
	return &RepoCreds{RepoUrl: make(map[string]string)}
}

func NewVCSCreds() *VCSCreds {
	return &VCSCreds{}
}

// these methods are to enable remoteconfig cred save with the proto VCSCreds object
func (m *VCSCreds) SetAcctNameAndType(name string, typ string) {
	m.AcctName = name
	m.Type = typ
}

func (m *VCSCreds) BuildCredPath(credType string, acctName string) string {
	return fmt.Sprintf("%s/vcs/%s/%s", "creds", acctName, credType)
}

func (m *VCSCreds) SetSecret(secret string) {
	m.ClientSecret = secret
}

func (m *VCSCreds) SetAdditionalFields(infoType string, val string) {
	switch infoType {
	case "clientid":
		m.ClientId = val
	case "tokenurl":
		m.TokenURL = val
	}
}

func (m *VCSCreds) AddAdditionalFields(consule *consul.Consulet, path string) error {
	err := consule.AddKeyValue(path+"/clientid", []byte(m.ClientId))
	if err != nil {
		return err
	}
	err = consule.AddKeyValue(path+"/tokenurl", []byte(m.TokenURL))
	if err != nil {
		return err
	}
	return err
}

func (m *VCSCreds) Spawn() cred.RemoteConfigCred {
	return &VCSCreds{}
}

// wrapper interface around models.BuildRuntimeInfo
type BuildRuntime interface {
	GetDone() bool
	GetIp() string
	GetGrpcPort() string
	GetHash() string
	CreateBuildClient() (protobuf.BuildClient, error)
}


// CreateBuildClient dials the grpc server at the werker endpoints
func (m *BuildRuntimeInfo) CreateBuildClient() (protobuf.BuildClient, error) {
	//TODO: this is insecure
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err :=  grpc.Dial(m.Ip + ":" + m.GrpcPort, opts...)
	if err != nil {
		return nil, err
	}
	return protobuf.NewBuildClient(conn), nil
}

func (pr *PollRequest) Validate() error {
	pr.Cron = strings.TrimSpace(pr.Cron)
	//todo: add validating acct/repo
	return nil
}