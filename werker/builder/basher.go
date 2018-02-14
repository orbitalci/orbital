package builder

import (
	"fmt"
	"bitbucket.org/level11consulting/ocelot/protos"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"strings"
)

const DefaultBitbucketURL = "https://x-token-auth:%s@bitbucket.org/%s.git"
const DefaultGithubURL = ""

type Basher struct {
	BbDownloadURL string
	GithubDownloadURL string
}

func (b *Basher) GetBbDownloadURL() string {
	if len(b.BbDownloadURL) > 0 {
		return b.BbDownloadURL
	}
	return DefaultBitbucketURL
}

func (b *Basher) GetGithubDownloadURL() string {
	if len(b.GithubDownloadURL) > 0 {
		return b.GithubDownloadURL
	}
	return DefaultGithubURL
}

func (b *Basher) SetBbDownloadURL(downloadURL string) {
	b.BbDownloadURL = downloadURL
}

func (b *Basher) SetGithubDownloadURL(downloadURL string) {
	b.GithubDownloadURL = downloadURL
}

//DownloadCodebase builds bash commands to be executed for downloading the codebase
func (b *Basher) DownloadCodebase(werk *protos.WerkerTask) []string {
	var downloadCode []string

	switch werk.VcsType {
	case "bitbucket":
		//if download url is not the default, then we assume whoever set it knows exactly what they're doing and no replacements
		if b.GetBbDownloadURL() != DefaultBitbucketURL {
			downloadCode = append(downloadCode, "/.ocelot/bb_download.sh", werk.VcsToken, b.GetBbDownloadURL(), werk.CheckoutHash)
		} else {
			downloadCode = append(downloadCode, "/.ocelot/bb_download.sh", werk.VcsToken, fmt.Sprintf(b.GetBbDownloadURL(), werk.VcsToken, werk.FullName), werk.CheckoutHash)
		}
	case "github":
		ocelog.Log().Error("not implemented")
	default:
		ocelog.Log().Error("werker VCS type not recognized")
	}

	return downloadCode
}

//DownloadSSHKey will using the vault token to try to download the ssh key located at the path + `/ssh`
func (b *Basher) DownloadSSHKey(vaultKey, vaultPath string) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf("/.ocelot/get_ssh_key.sh %s %s", vaultKey, vaultPath + "/ssh")}
}

func (b *Basher) WriteMavenSettingsXml(settingsXML string) []string {
	return []string{"/bin/sh", "-c", "/.ocelot/render_mvn.sh " + "'" + settingsXML + "'"}
}


//CDAndRunCmds will cd into the root directory of the codebase and execute commands passed in
func (b *Basher) CDAndRunCmds(cmds []string, commitHash string) []string {
	build := append([]string{"cd /" + commitHash}, cmds...)
	buildAndDeploy := append([]string{"/bin/sh", "-c", strings.Join(build, " && ")})
	return buildAndDeploy
}