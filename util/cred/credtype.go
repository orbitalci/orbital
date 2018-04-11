package cred

import (
	"fmt"
	"strings"

	pb "bitbucket.org/level11consulting/ocelot/admin/models"
)

// OcyCredType is the the of credential that we will be storing, ie binary repo or vcs
type OcyCredType int

const (
	Vcs OcyCredType = iota
	Repo
	K8s
)

var OcyCredMap = map[string]OcyCredType{
	"vcs": Vcs,
	"repo": Repo,
	"k8s": K8s,
}

func returnOcyStringCredType(sctype pb.SubCredType) string {
	return strings.ToLower(sctype.String())
}

func BuildCredPath(scType pb.SubCredType, AcctName string, ocyCredType pb.CredType) string {
	var pattern string
	switch ocyCredType {
	case pb.CredType_VCS: pattern = "%s/vcs/%s/%s"
	case pb.CredType_REPO: pattern = "%s/repo/%s/%s"
	case pb.CredType_K8S: pattern = "%s/k8s/%s/%s"
	default: panic("only repo|vcs|k8s")
	}
	return fmt.Sprintf(pattern, ConfigPath, AcctName, strings.ToLower(scType.String()))
}

// returns <vcs or repo>/acctname/credType/infoType
func splitConsulCredPath(path string) (typ OcyCredType, acctName, credType, infoType string) {
	pathKeys := strings.Split(path, "/")
	typ = OcyCredMap[pathKeys[1]]
	acctName = pathKeys[2]
	credType = pathKeys[3]
	infoType = strings.Join(pathKeys[4:], "/")

	return
}

