package buildruntime


import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"fmt"
	"github.com/hashicorp/consul/testutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)
// todo: put these in util
func modifyServerConfig(c *testutil.TestServerConfig) {
	c.LogLevel = "err"
}
// todo: put these in util
func initServerAndConsulet(t *testing.T) ( *consulet.Consulet, *testutil.TestServer, storage.BuildSum) {
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
//
//type ConsulShtuff struct {
//	c  *consulet.Consulet
//	ts *testutil.TestServer
//}

//SetBuildDone(consulete *consul.Consulet, gitHash string) error
func Test_SetBuildDone(t *testing.T) {
	hash := "OHAi"
	consu, serv, _ := initServerAndConsulet(t)
	defer serv.Stop()
	if err := SetBuildDone(consu, hash); err != nil {
		t.Fatal("could not set build done, err ", err.Error())
	}
	done := serv.GetKV(t, fmt.Sprintf(buildDonePath, hash))
	if string(done) != "true" {
		t.Error(test.StrFormatErrors("done flag", "true", string(done)))
	}
}

//// func CheckIfBuildDone(consulete *consul.Consulet, gitHash string) bool {
func Test_CheckIfBuildDone(t *testing.T) {
	hash := "sup"
	consu, serv, store := initServerAndConsulet(t)
	defer serv.Stop()
	testAddFullBuildSummary(hash, store, t)
	defer os.RemoveAll("./test-fixtures/storage")
	done := CheckIfBuildDone(consu, store, hash)
	if !done {
		t.Error(test.GenericStrFormatErrors("build done", true, done))
	}
	done = CheckIfBuildDone(consu, store, "nerd")
	if done {
		t.Error(test.GenericStrFormatErrors("build done", false, done))
	}
}

func testAddFullBuildSummary(hash string, store storage.BuildSum, t *testing.T) {
	id, err := store.AddSumStart(hash, time.Now(), "1", "2", "3")
	if err != nil {
		t.Fatal(err)
	}
	err = store.UpdateSum(false, 10.7, id)
	if err != nil {
		t.Fatal(err)
	}
}

// func Register(consulete *consul.Consulet, gitHash string, ip string, grpcPort string, wsPort string) (err error) {
func Test_Register(t *testing.T) {
	hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	wsPort := "4030"
	consu, serv, _ := initServerAndConsulet(t)
	defer serv.Stop()
	err := Register(consu, hash, ip, grpcPort, wsPort)
	if err != nil {
		t.Fatal("could not register with consul, err: ", err)
	}
	//t.Log("gettin done")
	//done := serv.GetKVString(t, fmt.Sprintf(buildDonePath, hash))
	//if done != "" {
	//	t.Error("should not have set done flag, set to: ", done)
	//}
	t.Log("gettin grpc")
	grp := serv.GetKVString(t, fmt.Sprintf(buildGrpcPort, hash))
	if grp != grpcPort {
		t.Error(test.StrFormatErrors("grpc port", grpcPort, grp))
	}
	t.Log("gettin wsp")
	wsP := serv.GetKVString(t, fmt.Sprintf(buildWsPort, hash))
	if wsP != wsPort {
		t.Error(test.StrFormatErrors("websocket port", wsPort, wsP))
	}
	t.Log("gettin ip")
	registeredIP := serv.GetKVString(t, fmt.Sprintf(buildRegister, hash))
	if registeredIP != ip {
		t.Error(test.StrFormatErrors("registered ip", ip, registeredIP))
	}

}
//
//// func GetBuildRuntime(consulete *consul.Consulet, gitHash string) (*BuildRuntime, error) {
func Test_GetBuildRuntime(t *testing.T) {
	hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	wsPort := "4030"
	consu, serv, _ := initServerAndConsulet(t)
	defer serv.Stop()
	serv.SetKV(t, fmt.Sprintf(buildGrpcPort, hash), []byte(grpcPort))
	serv.SetKV(t, fmt.Sprintf(buildWsPort, hash), []byte(wsPort))
	serv.SetKV(t, fmt.Sprintf(buildRegister, hash), []byte(ip))
	brt, err := GetBuildRuntime(consu, hash)
	if err != nil {
		t.Fatal("unable to get build runtime, err: ", err.Error())
	}
	if len(brt) != 1 {
		t.Error(test.GenericStrFormatErrors("result length", 1, len(brt)))
	}

	for _, val := range brt{
		if val.Done != false {
			t.Error(test.GenericStrFormatErrors("done", false, val.Done))
		}
		if val.GrpcPort != grpcPort {
			t.Error(test.StrFormatErrors("grpc port", grpcPort, val.GrpcPort))
		}
		if val.Ip != ip {
			t.Error(test.StrFormatErrors("registered ip", ip, val.Ip))
		}
	}

}