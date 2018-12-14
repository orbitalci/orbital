/*
workingdir.go handles everything related to generating a prefix for the .ocelot / clone directory
*/
package workingdir

import (
	"fmt"

	"github.com/level11consulting/orbitalci/models"
)

// GetOcyPrefixFromWerkerType will return "" for anything that runs in a container because root access can be assumed
// If it is running with the BARE connection (ie mac builds) or via Exec then it will find the home direc and use that as the prefix for the .ocelot directory
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
