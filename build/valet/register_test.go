package valet

import (
	"fmt"
	"testing"

	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/common"
	util "bitbucket.org/level11consulting/ocelot/common/testutil"

)


// todo: this belongs in valet
// func Register(consulete *consul.Consulet, gitHash string, ip string, grpcPort string, wsPort string) (err error) {
func Test_Register(t *testing.T) {
	//hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	wsPort := "4030"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	uuId, err := Register(consu, ip, grpcPort, wsPort)
	if err != nil {
		t.Fatal("could not register with consul, err: ", err)
	}
	//t.Log("gettin done")
	//done := serv.GetKVString(t, fmt.Sprintf(buildDonePath, hash))
	//if done != "" {
	//	t.Error("should not have set done flag, set to: ", done)
	//}
	t.Log("gettin grpc")
	grp := serv.GetKVString(t, fmt.Sprintf(common.WerkerGrpc, uuId.String()))
	if grp != grpcPort {
		t.Error(test.StrFormatErrors("grpc port", grpcPort, grp))
	}
	t.Log("gettin wsp")
	wsP := serv.GetKVString(t, fmt.Sprintf(common.WerkerWs, uuId.String()))
	if wsP != wsPort {
		t.Error(test.StrFormatErrors("websocket port", wsPort, wsP))
	}
	t.Log("gettin ip")
	registeredIP := serv.GetKVString(t, fmt.Sprintf(common.WerkerIp, uuId.String()))
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
	uuId, err := Register(consu, ip, grpcPort, wsPort)
	if err != nil {
		t.Fatal("could not register with consul, err: ", err)
	}
	if err = RegisterStartedBuild(consu, uuId.String(), hash); err != nil {
		t.Fatal("unable to register start of build, err: ", err.Error())
	}
	mapPath := common.MakeBuildMapPath(hash)
	if werkerId := serv.GetKVString(t, mapPath); werkerId != uuId.String() {
		t.Error("werker uuid", uuId.String(), werkerId)
	}
	if err = RegisterBuild(consu, uuId.String(), hash, dockerUuid); err != nil {
		t.Fatal("unable to register the build")
	}
	dockerUuidByte := serv.GetKV(t, common.MakeDockerUuidPath(uuId.String(), hash))
	returnedUuid := string(dockerUuidByte)
	if dockerUuid != returnedUuid {
		test.StrFormatErrors("docker uuid", dockerUuid, returnedUuid)
	}

}
//func Test_Delete(t *testing.T) {
//	werkerId := "werkerId"
//	hash := "1231231231"
//	dockerUuid := "12312324/81dfasd"
//	consu, serv := util.InitServerAndConsulet(t)
//	defer serv.Stop()
//	serv.SetKV(t, fmt.Sprintf(buildDockerUuid, werkerId, hash), []byte(dockerUuid))
//	serv.SetKV(t, fmt.Sprintf(werkerBuildMap, hash), []byte(werkerId))
//	if err := Delete(consu, hash); err != nil {
//		t.Fatal("could not delete!", err)
//	}
//	liveUuid, err := consu.GetKeyValue(fmt.Sprintf(buildDockerUuid, werkerId, hash))
//	if err != nil {
//		t.Fatal("unable to connect to consu ", err.Error())
//	}
//	if liveUuid != nil {
//		t.Error("liveUuid path should not exist after delete")
//	}
//	werkerIdd, err := consu.GetKeyValue(fmt.Sprintf(werkerBuildMap, hash))
//	if err != nil {
//		t.Fatal("unable to connect to conu ", err.Error())
//	}
//	if werkerIdd != nil {
//		t.Error("werkerId path should not exist after delete")
//	}
//
//}
//
