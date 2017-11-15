package consulet

import (
	"github.com/hashicorp/consul/testutil"
	"testing"
	"strings"
	"strconv"
)

func initServerAndConsulet(t *testing.T) (*Consulet, *testutil.TestServer) {
	testServer, err := testutil.NewTestServer()
	if err != nil {
		t.Fatal("Couldn't create consul test server, error: ", err)
	}
	ayy := strings.Split(testServer.HTTPAddr, ":")
	port, _ := strconv.ParseInt(ayy[1], 10, 32)
	consul := New(ayy[0], int(port))
	return consul, testServer
}

func TestConsulet_CreateNewSemaphore(t *testing.T) {
	consul, serv := initServerAndConsulet(t)
	defer serv.Stop()
	sema, err := consul.CreateNewSemaphore("test", 1)
	if err != nil {
		t.Errorf("couldnt create semaphore. error %s", err)
	}
	_, err = sema.Acquire(nil)
	if err !=  nil {
		t.Fatalf("could not acquire lock on first try, err %s", err)
	}
	_, err = sema.Acquire(nil)
	if err == nil {
		t.Fatal("should not be able to acquire twice.")
	}
	err = sema.Release()
	if err != nil {
		t.Fatalf("should be able to release lock. err %s", err)
	}
}