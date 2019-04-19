package buildmonitor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/common"
	util "github.com/level11consulting/ocelot/common/testutil"
)

func Test_Register(t *testing.T) {
	//hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	wsPort := "4030"
	tags := []string{"one", "two"}
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	uuId, err := Register(consu, ip, grpcPort, wsPort, tags)
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
	t.Log("gettin tags")
	registeredTags := serv.GetKVString(t, fmt.Sprintf(common.WerkerTags, uuId.String()))
	if diff := deep.Equal(strings.Split(registeredTags, ","), tags); diff != nil {
		t.Error(diff)
	}

}

func Test_RegisterBuild(t *testing.T) {
	hash := "1231231231"
	ip := "10.1.1.0"
	grpcPort := "1020"
	wsPort := "4030"
	dockerUuid := "1111-2222-3333-asdf"
	consu, serv := util.InitServerAndConsulet(t)
	defer serv.Stop()
	uuId, err := Register(consu, ip, grpcPort, wsPort, []string{})
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
