package admin

import (
	"github.com/level11consulting/ocelot/repo"
	"github.com/level11consulting/ocelot/secret"
	"github.com/level11consulting/ocelot/server/grpc/admin/status"
)

type OcelotServer interface {
	secret.SecretInterface
	repo.RepoInterface
	status.StatusInterface
	BuildInterface
}
