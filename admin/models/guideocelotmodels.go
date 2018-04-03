package models

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
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
	return cred.BuildCredPath(credType, acctName, cred.Repo)
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
	return cred.BuildCredPath(credType, acctName, cred.Vcs)
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

func NewK8sCreds() *K8SCreds {
	return &K8SCreds{}
}

func (m *K8SCreds) GetClientSecret() string {
	return m.K8SContents
}

func (m *K8SCreds) SetAcctNameAndType(name, typ string) {
	m.AcctName = name
	// no type here! mua ha ha. GetType() returns a dummy
}

func (m *K8SCreds) SetSecret(str string) {
	m.K8SContents = str
}

func (m *K8SCreds) SetAdditionalFields(key string, val string) {
	// do nothing, there is only one field.
}

func (m *K8SCreds) AddAdditionalFields(consule *consul.Consulet, path string) error {
	err := consule.AddKeyValue(path+"/exists", []byte("true"))
	return err
}

func (m *K8SCreds) BuildCredPath(credtype, acctName string) string {
	return cred.BuildCredPath(credtype, acctName, cred.K8s)
}

// todo: should we maybe use type to use different kubeconfigs? idk. for now there can only be 1
func (m *K8SCreds) GetType() string {
	return "k8s"
}

func (m *K8SCreds) Spawn() cred.RemoteConfigCred {
	return &K8SCreds{}
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