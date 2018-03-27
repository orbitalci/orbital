package util

import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	"github.com/hashicorp/consul/testutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func modifyServerConfig(c *testutil.TestServerConfig) {
	c.LogLevel = "err"
}

func InitServerAndConsulet(t *testing.T) ( *consulet.Consulet, *testutil.TestServer) {
	//todo: idk if we want to add the consul binary to our build image, but it is required
	BuildServerHack(t)
	testServer, err := testutil.NewTestServerConfig(modifyServerConfig)
	if err != nil {
		t.Fatal("Couldn't create consul test server, error: ", err)
	}
	ayy := strings.Split(testServer.HTTPAddr, ":")
	port, _ := strconv.ParseInt(ayy[1], 10, 32)
	consul, _ := consulet.New(ayy[0], int(port))
	return consul, testServer
}

//BuildServerHack will check the environment for a variable $BUILDSERVERHACK, will skip if it does not exist
func BuildServerHack(t *testing.T) {
	_, ok := os.LookupEnv("BUILDSERVERHACK")
	if ok {
		t.Skip("test flagged as build server hack, skipping.")
	}
}
