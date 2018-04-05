package nsqwatch

import (
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"github.com/hashicorp/consul/testutil"
	"net"
	"time"
	//"github.com/nsqio/go-nsq"
	"testing"
)


func getStructs(t *testing.T, store storage.OcelotStorage) (func(), net.Listener, *testutil.TestServer, *NsqWatch) {
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)
	consumer := nsqpb.NewDefaultProtoConsume()
	// don't care about errors right now because i don't actually want to connect
	consumer.ConsumeMessages("testtesttesttest", "test")
	nsqw := &NsqWatch{
		interval: 1,
		pConsumers: []*nsqpb.ProtoConsume{consumer},
		remoteConf: testRemoteConfig,
		store: store,
	}
	return func(){cred.TeardownVaultAndConsul(vaultListener, consulServer)}, vaultListener, consulServer, nsqw
}

func TestNsqWatch_MaintainHealths(t *testing.T) {
	rcHelathy := cred.NewHealthyMaintain()
	storeHealth := storage.NewHealthyStorage()
	consumer := nsqpb.NewDefaultProtoConsume()
	consumer.ConsumeMessages("testtesttesttest", "test")
	nsqw := &NsqWatch{
		interval: 1,
		pConsumers: []*nsqpb.ProtoConsume{consumer},
		remoteConf: rcHelathy,
		store: storeHealth,
	}
	go nsqw.MaintainHealths()
	if nsqw.paused {
		t.Error("everything is up, nsq consumer  should not be paused")

	}
	rcHelathy.IsHealthy = false
	rcHelathy.SuccessfulReconnect = false
	time.Sleep(2*time.Second)
	if !nsqw.paused {
		t.Error("vault has been shut down, nsq consumer  should be paused")
		return
	}
	rcHelathy.IsHealthy = true
	rcHelathy.SuccessfulReconnect = true
	time.Sleep(2*time.Second)
	if nsqw.paused {
		t.Error("everything is up, nsq consumer  should not be paused")
	}
	storeHealth.IsHealthy = false
	time.Sleep(2*time.Second)
	if !nsqw.paused {
		t.Error("postgres has been shut down, nsq consumer should be paused")
	}
}