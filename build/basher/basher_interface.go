package basher

import (
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

type OcyBash interface {
	GetBbDownloadURL() string
	GetGithubDownloadURL() string
	SetBbDownloadURL(downloadURL string)
	SetGithubDownloadURL(downloadURL string)
	DownloadCodebase(werk *pb.WerkerTask) []string
	DownloadSSHKey(vaultKey, vaultPath string) []string
	DownloadTemplateFiles(werkerPort string) []string
	DownloadKubectl(werkerPort string) []string
	CDAndRunCmds(cmds []string, commitHash string) []string
}