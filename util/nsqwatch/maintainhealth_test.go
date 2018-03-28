package nsqwatch

import (
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"time"

	//"github.com/nsqio/go-nsq"
	"testing"
)




func TestNsqWatch_MaintainHealths(t *testing.T) {
	cleanup, pw, port := storage.CreateTestPgDatabase(t)
	defer cleanup(t)
	pg := storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)

	consumer := nsqpb.NewDefaultProtoConsume()
	consumer.ConsumeMessages("testtesttesttest", "test")
	nsqw := &NsqWatch{
		interval: 1,
		pConsumers: []*nsqpb.ProtoConsume{consumer},
		remoteConf: testRemoteConfig,
		store: pg,
	}
	go nsqw.MaintainHealths()
	time.Sleep(1*time.Second)
	vaultListener.Close()
	time.Sleep(1*time.Second)
	if nsqw.paused != true {
		t.Error("vault has been shut down, nsq should be paused")
	}
	cred.TeardownVaultAndConsul(vaultListener, consulServer)
	// todo add checks for vault, consul
}