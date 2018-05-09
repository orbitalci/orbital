/*
prefix.go handles everything related to generating a prefix for the .ocelot / clone directory
*/
package build

import (
	"fmt"

	"github.com/shankj3/ocelot/models"
)

// GetOcyPrefixFromWerkerType will return "" for anything that runs in a container because root access can be assumed
// If it is running with the SSH connection (ie mac builds) or via Exec then it will find the home direc and use that as the prefix for the .ocelot directory
func GetOcyPrefixFromWerkerType(wt models.WerkType) string {
	switch wt {
	case models.SSH, models.Exec:
		return "/tmp"
	default:
		return ""
	}
}
//
func GetOcelotDir(prefix string) string {
	return prefix + "/.ocelot"
}

func GetPrefixDir(prefix string) string {
	return prefix
}

func GetCloneDir(prefix, hash string) string {
	return fmt.Sprintf("%s/%s", GetOcelotDir(prefix), hash)
}
