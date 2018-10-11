package basher

import (
	"errors"
	"fmt"
	"strings"

	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models/pb"
)

const DefaultBitbucketURL = "https://x-token-auth:%s@bitbucket.org/%s.git"
const DefaultGithubURL = ""

type Bashable func(string) []string

func NewBasher(downloadUrl, GHubDownloadUrl, LoopbackIP, dotOcelotPrefix string) (*Basher, error) {
	if dotOcelotPrefix != "" {
		// we are fine with using just slash, because all of this is bash/linux/unix specific
		if string(dotOcelotPrefix[0]) != "/" {
			return nil, errors.New("ocelot prefix must begin with a filepath separator")
		}
		if string(dotOcelotPrefix[len(dotOcelotPrefix)-1]) == "/" {
			dotOcelotPrefix = string(dotOcelotPrefix[(len(dotOcelotPrefix) - 1):])
		}
	}
	return &Basher{
		BbDownloadURL:     downloadUrl,
		GithubDownloadURL: GHubDownloadUrl,
		LoopbackIp:        LoopbackIP,
		ocelotPrefix:      dotOcelotPrefix,
	}, nil
}

type Basher struct {
	BbDownloadURL     string
	GithubDownloadURL string
	LoopbackIp        string
	ocelotPrefix      string
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

func (b *Basher) InstallPackageDeps() []string {
	return []string{"/bin/sh", "-c", b.OcelotDir() + "/install_deps.sh"}
}

//DownloadCodebase builds bash commands to be executed for downloading the codebase
func (b *Basher) DownloadCodebase(werk *pb.WerkerTask) []string {
	downloadCode := []string{"/bin/sh", "-c"}
	var downloadCmd string
	switch werk.VcsType {
	case pb.SubCredType_BITBUCKET:
		//if download url is not the default, then we assume whoever set it knows exactly what they're doing and no replacements
		if b.GetBbDownloadURL() != DefaultBitbucketURL {
			downloadCmd = fmt.Sprintf("%s/bb_download.sh %s %s %s %s", b.OcelotDir(), werk.VcsToken, b.GetBbDownloadURL(), werk.CheckoutHash, b.CloneDir(werk.CheckoutHash))
		} else {
			downloadCmd = fmt.Sprintf("%s/bb_download.sh %s %s %s %s", b.OcelotDir(), werk.VcsToken, fmt.Sprintf(b.GetBbDownloadURL(), werk.VcsToken, werk.FullName), werk.CheckoutHash, b.CloneDir(werk.CheckoutHash))
		}
	case pb.SubCredType_GITHUB:
		ocelog.Log().Error("not implemented")
	default:
		ocelog.Log().Error("werker VCS type not recognized")
	}

	downloadCode = append(downloadCode, downloadCmd)
	return downloadCode
}

//DownloadSSHKey will using the vault token to try to download the ssh key located at the path + `/ssh`
func (b *Basher) DownloadSSHKey(vaultKey, vaultPath string) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf("%s/get_ssh_key.sh %s %s", b.OcelotDir(), vaultKey, vaultPath+"/ssh")}
}

//DownloadTemplateFiles will download template files necessary to build containers from werker
func (b *Basher) DownloadTemplateFiles(werkerPort string) []string {
	////downloadLink := fmt.Sprintf("http://%s:%s/do_things.tar", b.LoopbackIp, werkerPort)
	//////downloadLink := fmt.Sprintf("http://172.17.0.1:%s/do_things.tar", werkerPort)
	//////warning: sleep has to be a integer; infinity doesn't exist everywhere.
	////command := fmt.Sprintf("mkdir %s && wget %s && tar -xf do_things.tar -C %s && cd %s  && chmod +x * && echo \"Ocelot has finished with downloading templates\" && sleep 3600", b.OcelotDir(), downloadLink, b.OcelotDir(), b.OcelotDir())
	sleepless := b.SleeplessDownloadTemplateFiles(b.LoopbackIp, werkerPort)
	command := sleepless + " && sleep 3600"
	return []string{"/bin/sh", "-c", command}
}

//SleeplessDownloadTemplateFiles will download the template files to
func (b *Basher) SleeplessDownloadTemplateFiles(werkerIp string, werkerPort string) string {
	downloadLink := fmt.Sprintf("http://%s:%s/do_things.tar", werkerIp, werkerPort)
	command := fmt.Sprintf("mkdir -p %s && wget %s && tar -xf do_things.tar -C %s && rm do_things.tar && cd %s  && chmod +x * && echo \"Ocelot has finished with downloading templates\"", b.OcelotDir(), downloadLink, b.OcelotDir(), b.OcelotDir())
	return command
}

func (b *Basher) DownloadKubectl(werkerPort string) []string {
	downloadLink := fmt.Sprintf("http://%s:%s/kubectl", b.LoopbackIp, werkerPort)
	return []string{"/bin/sh", "-c", "cd /bin && wget " + downloadLink + " && chmod +x kubectl"}
}

//CDAndRunCmds will cd into the root directory of the codebase and execute commands passed in
func (b *Basher) CDAndRunCmds(cmds []string, commitHash string) []string {
	cdCmd := fmt.Sprintf("cd %s", b.CloneDir(commitHash))
	bild := append([]string{cdCmd}, cmds...)
	buildAndDeploy := append([]string{"/dev/init", "-s", "--", "/bin/sh", "-c", strings.Join(bild, " && ")})
	return buildAndDeploy
}

func (b *Basher) OcelotDir() string {
	return build.GetOcelotDir(b.ocelotPrefix)
}

func (b *Basher) PrefixDir() string {
	return build.GetPrefixDir(b.ocelotPrefix)
}

func (b *Basher) CloneDir(hash string) string {
	return build.GetCloneDir(b.ocelotPrefix, hash)
}
