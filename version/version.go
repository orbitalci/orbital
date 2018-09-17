package version

import (
	"fmt"
	"os"
	"strings"
)

var (
	// The git commit that was compiled. These will be filled in by the
	// compiler.
	GitCommit   string
	GitDescribe string

	// The main version number that is being run at the moment.
	Version = "0.2.1"

	// A pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease = "dev"
)

// GetHumanVersion composes the parts of the version in a way that's suitable
// for displaying to humans.
func GetHumanVersion() string {
	version := Version
	if GitDescribe != "" {
		version = GitDescribe
	}

	release := VersionPrerelease
	if GitDescribe == "" && release == "" {
		release = "dev"
	}
	if release != "" {
		version += fmt.Sprintf("-%s", release)
		if GitCommit != "" {
			version += fmt.Sprintf(" (%s)", GitCommit)
		}
	}

	// Strip off any single quotes added by the git information.
	return strings.Replace(version, "'", "", -1)
}

func GetShort() string {
	version := Version
	if GitDescribe != "" {
		version = GitDescribe
	}

	release := VersionPrerelease
	if GitDescribe == "" && release == "" {
		release = "dev"
	}
	if release != "" {
		version += fmt.Sprintf("-%s", release)
	}

	// Strip off any single quotes added by the git information.
	return strings.Replace(version, "'", "", -1)
}

//MaybePrintVersion super simple get for version in remainder of flag arguments.
// it will, if the length of the leftover arguments > 0; check if the any of the args is "version"
// if it is, it will print the version and exit
func MaybePrintVersion(remainders []string) {
	switch len(remainders) {
	case 0:
		return
	default:
		for _, remain := range remainders {
			if remain == "version" {
				fmt.Println(GetHumanVersion())
				os.Exit(0)
			}
		}
	}

}
