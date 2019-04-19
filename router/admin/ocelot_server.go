package admin

import (
	"github.com/level11consulting/ocelot/secret"
)

type OcelotServer interface {
	secret.SecretInterface
	RepoInterface
	StatusInterface
	BuildInterface
}
