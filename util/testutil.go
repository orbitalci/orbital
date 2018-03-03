package util

import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"github.com/hashicorp/consul/testutil"
	"strconv"
	"strings"
	"testing"
)

func modifyServerConfig(c *testutil.TestServerConfig) {
	c.LogLevel = "err"
}

func InitServerAndConsulet(t *testing.T) ( *consulet.Consulet, *testutil.TestServer, storage.BuildSum) {
	testServer, err := testutil.NewTestServerConfig(modifyServerConfig)
	if err != nil {
		t.Fatal("Couldn't create consul test server, error: ", err)
	}
	ayy := strings.Split(testServer.HTTPAddr, ":")
	port, _ := strconv.ParseInt(ayy[1], 10, 32)
	consul, _ := consulet.New(ayy[0], int(port))
	store := storage.NewFileBuildStorage("./test-fixtures/storage")
	return consul, testServer, store
}
