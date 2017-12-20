package builder

import (
	"fmt"
	"bitbucket.org/level11consulting/ocelot/protos"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"strings"
)

// TODO: don't think we need a struct for this? Maybe not

func DownloadCodebase(werk *protos.WerkerTask) []string {
	var downloadCode []string

	switch werk.VcsType {
	case "bitbucket":
		downloadCode = append(downloadCode, ".ocelot/bb_download.sh", werk.VcsToken, fmt.Sprintf("https://bitbucket.org/%s/get", werk.FullName), werk.CheckoutHash)
	case "github":
		ocelog.Log().Error("not implemented")
	}

	return downloadCode
}

func BuildAndDeploy(cmds []string, commitHash string) []string {
	build := append([]string{"cd /" + commitHash}, cmds...)
	buildAndDeploy := append([]string{"/bin/sh", "-c", strings.Join(build, " && ")})
	return buildAndDeploy
}