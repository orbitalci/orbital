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
	if testing.Short() {
		t.Skip("skipping MaintainHealths test because requires docker, and multiple consul restarts and that is slooow")
	}
	cleanup, pw, port := storage.CreateTestPgDatabase(t)
	// don't defer because it'll fail..
	//defer cleanup(t)
	pg := storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	cleanCV, vaultListener, _, nsqw := getStructs(t, pg)
	go nsqw.MaintainHealths()
	if nsqw.paused {
		t.Error("everything is up, nsq consumer  should not be paused")
	}
	time.Sleep(1*time.Second)
	vaultListener.Close()
	time.Sleep(1*time.Second)
	if !nsqw.paused {
		t.Fatal("vault has been shut down, nsq consumer  should be paused")
	}
	cleanCV()

	cleanCV, _, consulServer, nsqw := getStructs(t, pg)
	go nsqw.MaintainHealths()
	time.Sleep(1*time.Second)
	consulServer.Stop()
	time.Sleep(3*time.Second)
	if !nsqw.paused {
		t.Fatal("consul has been shut down, nsq consumer should be paused")
	}
	cleanCV()
	cleanCV, _, _, nsqw = getStructs(t, pg)
	go nsqw.MaintainHealths()
	time.Sleep(1*time.Second)
	cleanup(t)
	time.Sleep(2*time.Second)
	if !nsqw.paused {
		t.Fatal("postgres has been shut down, nsq consumer should be paused")
	}

}