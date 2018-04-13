package models

import (
	"errors"
	"strings"
)

var (
	vcsSubTypes = []SubCredType{SubCredType_BITBUCKET, SubCredType_GITHUB}
	repoSubTypes = []SubCredType{SubCredType_NEXUS, SubCredType_MAVEN, SubCredType_DOCKER}
	k8sSubTypes = []SubCredType{SubCredType_KUBECONF}
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

//
//func (this Stuff) UnmarshalJSON(b []byte) error {
//	var stuff map[string]string
//	err := json.Unmarshal(b, &stuff)
//	if err != nil {
//		return err
//	}
//	for key, value := range stuff {
//		numericKey, err := strconv.Atoi(key)
//		if err != nil {
//			return err
//		}
//		this[numericKey] = value
//	}
//	return nil
//}
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