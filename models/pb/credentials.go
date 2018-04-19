package pb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

//OcyCredder is an interface for interacting with credentials in Ocelot
type OcyCredder interface {
	SetSecret(string)
	UnmarshalAdditionalFields(fields []byte) error
	CreateAdditionalFields() ([]byte, error)
	GetClientSecret() string
	GetAcctName() string
	GetIdentifier() string
	GetSubType() SubCredType
	ValidateForInsert() *ValidationErr
}

func Invalidate(reason string) *ValidationErr {
	return &ValidationErr{msg: reason}
}

type ValidationErr struct {
	msg string
}

func (v *ValidationErr) Error() string {
	return v.msg
}

func NewRepoCreds() *RepoCreds {
	return &RepoCreds{}
}


func (m *RepoCreds) SetSecret(secret string) {
	m.Password = secret
}

func (m *RepoCreds) GetClientSecret() string {
	return m.Password
}


func (m *RepoCreds) CreateAdditionalFields() ([]byte, error) {
	fields := make(map[string]string)
	fields["username"] = m.Username
	fields["url"] = m.RepoUrl
	bytes, err := json.Marshal(fields)
	return bytes, err
}

func (m *RepoCreds) UnmarshalAdditionalFields(fields []byte) error {
	unmarshaled := make(map[string]string)
	if err := json.Unmarshal(fields, &unmarshaled); err != nil {
		return err
	}
	var ok bool
	if m.RepoUrl, ok = unmarshaled["url"]; !ok {
		return errors.New(fmt.Sprintf("repo url was not in field map, map is %v", unmarshaled))
	}
	if m.Username, ok = unmarshaled["username"]; !ok {
		return errors.New(fmt.Sprintf("username was not in field map, map is %v", unmarshaled))
	}
	return nil
}

func (m *RepoCreds) ValidateForInsert() *ValidationErr {
	errr := validateCommonFieldsForInsert(m)
	if m.RepoUrl == "" {
		errr = append(errr, "repoUrl is required")
	}
	if m.Username == "" {
		errr = append(errr, "username is required")
	}
	if len(errr) != 0 {
		return Invalidate(strings.Join(errr, "\n"))
	}
	return nil
}


func NewVCSCreds() *VCSCreds {
	return &VCSCreds{}
}


func (m *VCSCreds) CreateAdditionalFields() ([]byte, error) {
	fields := make(map[string]string)
	fields["tokenUrl"] = m.TokenURL
	fields["clientId"] = m.ClientId
	bytes, err := json.Marshal(fields)
	return bytes, err
}

func (m *VCSCreds) UnmarshalAdditionalFields(fields []byte) error {
	unmarshaled := make(map[string]string)
	if err := json.Unmarshal(fields, &unmarshaled); err != nil {
		return err
	}
	var ok bool
	if m.TokenURL, ok = unmarshaled["tokenUrl"]; !ok {
		return errors.New(fmt.Sprintf("token url was not in field map, map is %v", unmarshaled))
	}
	if m.ClientId, ok = unmarshaled["clientId"]; !ok {
		return errors.New(fmt.Sprintf("client id was not in field map, map is %v", unmarshaled))
	}
	return nil
}

func (m *VCSCreds) SetSecret(sec string) {
	m.ClientSecret = sec
}

// identifier for vcs creds will always be "<BITBUCKET|GITHUB|..>/<ACCTNAME>"
func (m *VCSCreds) BuildIdentifier() string {
	// can ignore error here, because VcsCreds will always have subtype in VCS
	identifier, _ := CreateVCSIdentifier(m.SubType, m.AcctName)
	return identifier
}

func (m *VCSCreds) ValidateForInsert() *ValidationErr {
	errr := validateCommonFieldsForInsert(m)
	if m.ClientId == "" {
		errr = append(errr, "oauth client id is required")
	}
	if m.TokenURL == "" {
		errr = append(errr, "oauth token url is required")
	}
	if len(errr) != 0 {
		return Invalidate(strings.Join(errr, "\n"))
	}
	return nil
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


func (m *K8SCreds) CreateAdditionalFields() ([]byte, error) {
	return []byte("{}"), nil
}

func (m *K8SCreds) UnmarshalAdditionalFields(fields []byte) error {
	return nil
}


func (m *K8SCreds) ValidateForInsert() *ValidationErr {
	errr := validateCommonFieldsForInsert(m)
	if len(errr) != 0 {
		return Invalidate(strings.Join(errr, "\n"))
	}
	return nil
}

func (m *SSHKeyWrapper) GetClientSecret() string {
	return string(m.PrivateKey)
}

func (m *SSHKeyWrapper) SetSecret(str string) {
	m.PrivateKey = []byte(str)
}

func (m *SSHKeyWrapper) CreateAdditionalFields() ([]byte, error) {
	return []byte("{}"), nil
}

func (m *SSHKeyWrapper) UnmarshalAdditionalFields(fields []byte) error {
	return nil
}

func (m *SSHKeyWrapper) ValidateForInsert() *ValidationErr {
	errr := validateCommonFieldsForInsert(m)
	if len(errr) != 0 {
		return Invalidate(strings.Join(errr, "\n"))
	}
	return nil
}


func validateCommonFieldsForInsert(credder OcyCredder) (errors []string) {
	if credder.GetIdentifier() == "" {
		errors = append(errors, "identifier is required, creds need a unique name to identify by")
	}
	if credder.GetAcctName() == "" {
		errors = append(errors, "account name is required")
	}
	if credder.GetClientSecret() == "" {
		errors = append(errors, "client secret is required")
	}
	if credder.GetSubType() == SubCredType_NIL_SCT {
		errors = append(errors, "subtype not instantiated")
	}
	return
}


// wrapper interface around models.BuildRuntimeInfo
type BuildRuntime interface {
	GetDone() bool
	GetIp() string
	GetGrpcPort() string
	GetHash() string
	CreateBuildClient() (BuildClient, error)
}



func (pr *PollRequest) Validate() error {
	pr.Cron = strings.TrimSpace(pr.Cron)
	//todo: add validating acct/repo
	return nil
}

var (
	vcsSubTypes = []SubCredType{SubCredType_BITBUCKET, SubCredType_GITHUB}
	repoSubTypes = []SubCredType{SubCredType_NEXUS, SubCredType_MAVEN, SubCredType_DOCKER}
	k8sSubTypes = []SubCredType{SubCredType_KUBECONF}
	sshSubTypes = []SubCredType{SubCredType_SSHKEY}
)

// Subtypes will return all the SubCredTypes that are associated with that CredType. Will return nil if it is unknown
func (x CredType) Subtypes() []SubCredType {
	switch x {
	case CredType_VCS:
		return vcsSubTypes
	case CredType_REPO:
		return repoSubTypes
	case CredType_K8S:
		return k8sSubTypes
	case CredType_SSH:
		return sshSubTypes
	}
	// this shouldn't happen, unless a new CredType is added and not updated here.
	return nil
}

func (x CredType) SubtypesString() []string {
	var subtypes []string
	switch x {
	case CredType_VCS:
		for _, st := range vcsSubTypes {
			subtypes = append(subtypes, st.String())
		}
	case CredType_REPO:
		for _, st := range repoSubTypes {
			subtypes = append(subtypes, st.String())
		}
	case CredType_K8S:
		for _, st := range repoSubTypes {
			subtypes = append(subtypes, st.String())
		}
	case CredType_SSH:
		for _, st := range sshSubTypes {
			subtypes = append(subtypes, st.String())
		}
	}
	return subtypes
}

//SpawnCredStruct will instantiate an Cred object with account, identifier, subcredtype, and credtype
func (x CredType) SpawnCredStruct(account, identifier string, subCredType SubCredType) OcyCredder {
	switch x {
	case CredType_VCS:
		return &VCSCreds{AcctName: account, Identifier: identifier, SubType: subCredType}
	case CredType_REPO:
		return &RepoCreds{AcctName: account, Identifier: identifier, SubType: subCredType}
	case CredType_K8S:
		return &K8SCreds{AcctName: account, Identifier: identifier, SubType: subCredType}
	case CredType_SSH:
		return &SSHKeyWrapper{AcctName: account, Identifier: identifier, SubType:subCredType}
	default:
		return nil
	}
}

func (x SubCredType) Parent() CredType {
	switch {
	case Contains(x, k8sSubTypes):
		return CredType_K8S
	case Contains(x, vcsSubTypes):
		return CredType_VCS
	case Contains(x, repoSubTypes):
		return CredType_REPO
	case Contains(x, sshSubTypes):
		return CredType_SSH
	}
	return -1
}

func Contains(credType SubCredType, types []SubCredType) bool {
	for _, typ := range types {
		if credType == typ {
			return true
		}
	}
	return false
}

func (i *SubCredType) MarshalJSON() ([]byte, error) {
	return []byte(strings.ToLower(i.String())), nil
}

func (i *SubCredType) UnmarshalJSON(b []byte) error {
	name := string(b)
	typ, ok := SubCredType_value[name]
	if !ok {
		return errors.New("not in subcredtype map")
	}
	*i = SubCredType(typ)
	return nil
}

// MarshalYAML implements a YAML Marshaler for SubCredType
func (i SubCredType) MarshalYAML() (interface{}, error) {
	return i.String(), nil
}

// UnmarshalYAML implements a YAML Unmarshaler for SubCredType
func (i *SubCredType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	var err error
	sct, ok := SubCredType_value[strings.ToUpper(s)]
	if !ok {
		return errors.New("not found in SubCredType_value map")
	}
	*i = SubCredType(sct)
	return err
}

func CreateVCSIdentifier(sct SubCredType, acctName string) (string, error){
	if !Contains(sct, CredType_VCS.Subtypes()) {
		return "", errors.New("must be of type CredType_VCS")
	}
	identifier := SubCredType_name[int32(sct)] + "/" + acctName
	return identifier, nil
}