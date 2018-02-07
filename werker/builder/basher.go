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

func (b *Basher) WriteMavenSettingsXml(settingsXML string) []string {
	return []string {"/bin/sh", "-c", "/.ocelot/render_mvn.sh " + settingsXML}
}

func (b *Basher) BuildScript(cmds []string, commitHash string) []string {
	build := append([]string{"cd /" + commitHash}, cmds...)
	buildAndDeploy := append([]string{"/bin/sh", "-c", strings.Join(build, " && ")})
	return buildAndDeploy
}

func (b *Basher) PushToNexus(commitHash string) []string {
	push := []string{"cd /" + commitHash}
	//TODO: how to tell if generated artifact is jar? What if they generate other artifact types?
	//mvnCmd := "mvn deploy:deploy-file -DgeneratePom=false -Dpackaging=jar -DrepositoryId=nexus -Durl=http://52.26.105.112:8081/nexus/content/repositories/snapshots -Dfile=/home/mariannefeng/git/test/test-ocelot/target/exampleboot-0.0.1-SNAPSHOT.jar -DpomFile=/home/mariannefeng/git/test/test-ocelot/pom.xml"
	//push = append(push, mvnCmd)
	runPush := append([]string{"/bin/sh", "-c", strings.Join(push, " && ")})
	return runPush
}