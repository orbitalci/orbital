package admin

type OcelotServer interface {
	SecretInterface
	RepoInterface
	StatusInterface
	BuildInterface
}
