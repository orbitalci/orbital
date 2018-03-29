package config

import (
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/util"
	"os"
	"testing"
)

// ** this test will pass only if vault token is set as env. variable **
// i'm really testing namsrals flag code, i dont trust it.
func TestGetConf_fromEnv(t *testing.T) {
	util.BuildServerHack(t)
	testConf := &WerkerConf{
		ServicePort: "9899",
		GrpcPort:    defaultGrpcPort,
		WerkerName:  "oh_YEEEAH",
		WerkerType:  Kubernetes,
		LogLevel:    "error",
		RegisterIP:  "55.259.12.197",
	}
	os.Setenv("WS_PORT", testConf.ServicePort)
	os.Setenv("TYPE", "kubernetes")
	os.Setenv("NAME", testConf.WerkerName)
	os.Setenv("LOG_LEVEL", testConf.LogLevel)
	os.Setenv("REGISTER_IP", testConf.RegisterIP)
	conf, err := GetConf()
	if err != nil {
		t.Fatal("no go ", err)
	}

	if conf.WerkerType != testConf.WerkerType {
		t.Error(test.GenericStrFormatErrors("werker type", testConf.WerkerType, conf.WerkerType))
	}
	if conf.ServicePort != testConf.ServicePort {
		t.Error(test.StrFormatErrors("service ws port", testConf.ServicePort, conf.ServicePort))
	}
	if conf.GrpcPort != testConf.GrpcPort {
		t.Error(test.StrFormatErrors("grpc port", testConf.GrpcPort, conf.GrpcPort))
	}
	if conf.WerkerName != testConf.WerkerName {
		t.Error(test.StrFormatErrors("werker name", testConf.WerkerName, conf.WerkerName))
	}
	if conf.LogLevel != testConf.LogLevel {
		t.Error(test.StrFormatErrors("log level", testConf.LogLevel, conf.LogLevel))
	}
	if conf.RegisterIP != testConf.RegisterIP {
		t.Error(test.StrFormatErrors("register ip", testConf.RegisterIP, conf.RegisterIP))
	}

	if conf.WerkerType != Kubernetes {
		t.Error("whuy doooo")
	}
}