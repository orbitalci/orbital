package models

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"fmt"
	"google.golang.org/grpc"
)


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
	switch infoType {
	case "repourl":
		m.RepoUrl = val
	case "username":
		m.Username = val
	}
}

func (m *RepoCreds) AddAdditionalFields(consule *consul.Consulet, path string) (err error) {
	if err := consule.AddKeyValue(path + "/username", []byte(m.Username)); err != nil {
		return err
	}
	if err = consule.AddKeyValue(path + "/repourl", []byte(m.RepoUrl)); err != nil {
		return err
	}
	return err
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

// wrapper interface around models.BuildRuntimeInfo
type BuildRuntime interface {
	GetDone() bool
	GetIp() string
	GetGrpcPort() string
	CreateBuildClient(opts []grpc.DialOption) (protobuf.BuildClient, error)
}


// CreateBuildClient dials the grpc server at the werker endpoints
func (m *BuildRuntimeInfo) CreateBuildClient(opts []grpc.DialOption) (protobuf.BuildClient, error) {
	conn, err :=  grpc.Dial(m.Ip + ":" + m.GrpcPort, opts...)
	if err != nil {
		return nil, err
	}
	return protobuf.NewBuildClient(conn), nil
}