package valet

//import (
//	"testing"
//	"bitbucket.org/level11consulting/ocelot/build"
//	util "bitbucket.org/level11consulting/ocelot/common/testutil"
//	"fmt"
//
//)
//
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
