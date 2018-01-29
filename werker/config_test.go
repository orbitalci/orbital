package werker

import (
	"bitbucket.org/level11consulting/go-til/test"
	"os"
	"testing"
)

// ** this test will pass only if vault token is set as env. variable **
// i'm really testing namsrals flag code, i dont trust it.
func TestGetConf_fromEnv(t *testing.T) {
	testConf := &WerkerConf{
		servicePort: "9899",
		grpcPort: defaultGrpcPort,
		WerkerName: "oh_YEEEAH",
		werkerType: Kubernetes,
		LogLevel: "error",
		RegisterIP: "55.259.12.197",
	}
	os.Setenv("WERKER_WS_PORT", testConf.servicePort)
	os.Setenv("WERKER_TYPE", "kubernetes")
	os.Setenv("WERKER_NAME", testConf.WerkerName)
	os.Setenv("WERKER_LOG_LEVEL", testConf.LogLevel)
	os.Setenv("WERKER_REGISTER_IP", testConf.RegisterIP)
	conf, err := GetConf()
	if err != nil {
		t.Fatal("no go ", err)
	}

	if conf.werkerType != testConf.werkerType {
		t.Error(test.GenericStrFormatErrors("werker type", testConf.werkerType, conf.werkerType))
	}
	if conf.servicePort != testConf.servicePort {
		t.Error(test.StrFormatErrors("service ws port", testConf.servicePort, conf.servicePort))
	}
	if conf.grpcPort != testConf.grpcPort {
		t.Error(test.StrFormatErrors("grpc port", testConf.grpcPort, conf.grpcPort))
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

	if conf.werkerType != Kubernetes {
		t.Error("whuy doooo")
	}
}