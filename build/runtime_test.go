package build


import (
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/build/valet"
	util "bitbucket.org/level11consulting/ocelot/common/testutil"
	"bitbucket.org/level11consulting/ocelot/storage"
	"fmt"
	"os"
	"testing"
)

//// func CheckIfBuildDone(consulete *consul.Consulet, gitHash string) bool {
func Test_CheckIfBuildDone(t *testing.T) {
	hash := "sup"
	consu, serv := util.InitServerAndConsulet(t)
	store := storage.CreateTestFileSystemStorage(t)
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
	id, err := store.AddSumStart(hash, "1", "2", "3")
	if err != nil {
		t.Fatal(err)
	}
	err = store.UpdateSum(false, 10.7, id)
	if err != nil {
		t.Fatal(err)
	}
}
// todo: this belongs in valet
// func Register(consulete *consul.Consulet, gitHash string, ip string, grpcPort string, wsPort string) (err error) {
func Test_Register(t *testing.T) {
	//hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	wsPort := "4030"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	uuid, err := valet.Register(consu, ip, grpcPort, wsPort)
	if err != nil {
		t.Fatal("could not register with consul, err: ", err)
	}
	//t.Log("gettin done")
	//done := serv.GetKVString(t, fmt.Sprintf(buildDonePath, hash))
	//if done != "" {
	//	t.Error("should not have set done flag, set to: ", done)
	//}
	t.Log("gettin grpc")
	grp := serv.GetKVString(t, fmt.Sprintf(werkerGrpc, uuid.String()))
	if grp != grpcPort {
		t.Error(test.StrFormatErrors("grpc port", grpcPort, grp))
	}
	t.Log("gettin wsp")
	wsP := serv.GetKVString(t, fmt.Sprintf(werkerWs, uuid.String()))
	if wsP != wsPort {
		t.Error(test.StrFormatErrors("websocket port", wsPort, wsP))
	}
	t.Log("gettin ip")
	registeredIP := serv.GetKVString(t, fmt.Sprintf(werkerIp, uuid.String()))
	if registeredIP != ip {
		t.Error(test.StrFormatErrors("registered ip", ip, registeredIP))
	}

}

// todo: this actually belongs in valet
func Test_RegisterBuild(t *testing.T) {
	hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	wsPort := "4030"
	dockerUuid := "1111-2222-3333-asdf"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	uuid, err := valet.Register(consu, ip, grpcPort, wsPort)
	if err != nil {
		t.Fatal("could not register with consul, err: ", err)
	}
	if err = valet.RegisterStartedBuild(consu, uuid.String(), hash); err != nil {
		t.Fatal("unable to register start of build, err: ", err.Error())
	}
	mapPath := MakeBuildMapPath(hash)
	if werkerId := serv.GetKVString(t, mapPath); werkerId != uuid.String() {
		t.Error("werker uuid", uuid.String(), werkerId)
	}
	if err = valet.RegisterBuild(consu, uuid.String(), hash, dockerUuid); err != nil {
		t.Fatal("unable to register the build")
	}
	dockerUuidByte := serv.GetKV(t, MakeDockerUuidPath(uuid.String(), hash))
	returnedUuid := string(dockerUuidByte)
	if dockerUuid != returnedUuid {
		test.StrFormatErrors("docker uuid", dockerUuid, returnedUuid)
	}

}

//
//// func GetBuildRuntime(consulete *consul.Consulet, gitHash string) (*BuildRuntime, error) {
func Test_GetBuildRuntime(t *testing.T) {
	hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	werkerId := "werkerId"
	wsPort := "4030"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	serv.SetKV(t, fmt.Sprintf(werkerGrpc, werkerId), []byte(grpcPort))
	serv.SetKV(t, fmt.Sprintf(werkerWs, werkerId), []byte(wsPort))
	serv.SetKV(t, fmt.Sprintf(werkerIp, werkerId), []byte(ip))
	serv.SetKV(t, fmt.Sprintf(buildDockerUuid, werkerId, hash), []byte("dockerId"))
	serv.SetKV(t, fmt.Sprintf(werkerBuildMap, hash), []byte(werkerId))
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

// todo: this also belongs in valet
func Test_Delete(t *testing.T) {
	werkerId := "werkerId"
	hash := "1231231231"
	dockerUuid := "12312324/81dfasd"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	serv.SetKV(t, fmt.Sprintf(buildDockerUuid, werkerId, hash), []byte(dockerUuid))
	serv.SetKV(t, fmt.Sprintf(werkerBuildMap, hash), []byte(werkerId))
	if err := Delete(consu, hash); err != nil {
		t.Fatal("could not delete!", err)
	}
	liveUuid, err := consu.GetKeyValue(fmt.Sprintf(buildDockerUuid, werkerId, hash))
	if err != nil {
		t.Fatal("unable to connect to consu ", err.Error())
	}
	if liveUuid != nil {
		t.Error("liveUuid path should not exist after delete")
	}
	werkerIdd, err := consu.GetKeyValue(fmt.Sprintf(werkerBuildMap, hash))
	if err != nil {
		t.Fatal("unable to connect to conu ", err.Error())
	}
	if werkerIdd != nil {
		t.Error("werkerId path should not exist after delete")
	}

}
