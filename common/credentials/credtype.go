package credentials

import (
	"fmt"
	"strings"

	"bitbucket.org/level11consulting/ocelot/common"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)


//BuildCred path will return "<creds>/vcs|repo|k8s/<subCredType(string)>/<identifier>"
func BuildCredPath(scType pb.SubCredType, AcctName string, ocyCredType pb.CredType, identifier string) string {
	var pattern string
	switch ocyCredType {
	case pb.CredType_VCS: pattern = "%s/vcs/%s/%s/%s"
	case pb.CredType_REPO: pattern = "%s/repo/%s/%s/%s"
	case pb.CredType_K8S: pattern = "%s/k8s/%s/%s/%s"
	//this will not happen in real life because all the setcred methods for guideOcelotServer check for this specific issue
	default: panic("only repo|vcs|k8s")
	}
	path := fmt.Sprintf(pattern, common.ConfigPath, AcctName, strings.ToLower(scType.String()), identifier)
	return path
}
