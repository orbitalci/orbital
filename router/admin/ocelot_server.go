package admin

import (
	"github.com/level11consulting/ocelot/repo"
	"github.com/level11consulting/ocelot/secret"
)

type OcelotServer interface {
	secret.SecretInterface
	repo.RepoInterface
	StatusInterface
	BuildInterface
}
