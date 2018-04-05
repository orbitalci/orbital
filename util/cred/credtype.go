package cred

import (
	"fmt"
	"strings"
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


func BuildCredPath(credType string, AcctName string, ocyCredType OcyCredType) string {
	var pattern string
	switch ocyCredType {
	case Vcs: pattern = "%s/vcs/%s/%s"
	case Repo: pattern = "%s/repo/%s/%s"
	case K8s: pattern = "%s/k8s/%s/%s"
	default: panic("only repo|vcs|k8s")
	}
	return fmt.Sprintf(pattern, ConfigPath, AcctName, credType)
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

