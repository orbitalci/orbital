package builder

import (
	"bitbucket.org/level11consulting/ocelot/old/protos"
)

type OcyBash interface {
	GetBbDownloadURL() string
	GetGithubDownloadURL() string
	SetBbDownloadURL(downloadURL string)
	SetGithubDownloadURL(downloadURL string)
	DownloadCodebase(werk *protos.WerkerTask) []string
	DownloadSSHKey(vaultKey, vaultPath string) []string
	WriteMavenSettingsXml(settingsXML string) []string
	WriteDockerJson(encodedDocker string) []string
	DownloadTemplateFiles(werkerPort string) []string
	DownloadKubectl(werkerPort string) []string
	InstallKubeconfig(encodedKubeConf string) []string
	CDAndRunCmds(cmds []string, commitHash string) []string
}