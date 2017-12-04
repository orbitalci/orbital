package werker

import (
	//"fmt"
	"testing"
	"github.com/shankj3/ocelot/util"
	"bytes"
	"github.com/shankj3/ocelot/protos/out"
	"github.com/golang/protobuf/proto"
)

func TestWorkerMsgHandler_WatchForResults(t *testing.T) {
	var watchdata = []struct {
		name string
		hash string
		chanData []byte
	}{
		{"ice","hash hash baby", []byte("ice ice baby")},
		{"bean", "pinto bean", []byte("black bean")},
		{"fruit", "jackfruit", []byte("idk.. whats like a jackfruit but *not* a jackfruit?")},
	}
	wmh := testGetWorkerMsgHandler(t, "test wfr")
	for _, wd := range watchdata {
		wd := wd
		t.Run(wd.name, func(t *testing.T){
			go wmh.WatchForResults(wd.hash)
			go func(){
				wmh.infochan <- wd.chanData
			}()
			trans := <- wmh.ChanChan
			info := <- trans.InfoChan
			if bytes.Compare(info, wd.chanData) != 0 {
				t.Error(util.StrFormatErrors("info channel response", string(wd.chanData), string(info)))
			}
			if wd.hash != trans.Hash {
				t.Error(util.StrFormatErrors("git hash", wd.hash, trans.Hash))
			}
		})
	}
}

func TestWorkerMsgHandler_buildPRBuildBundle(t *testing.T) {
	msg := &protos.PRBuildBundle{
		CheckoutHash: "1231231231234",
	}
	wmh := testGetWorkerMsgHandler(t, "testbuild")
	go wmh.build(msg)
	// todo: why does data := <-wmh.infochan need to be inside a goroutine, but not this wmh.ChanChan???
	w, ok := <-wmh.ChanChan; if !ok {
		t.Fatal("no data on channel")
	} else {
		if w.Hash != msg.CheckoutHash {
			t.Error(util.StrFormatErrors("checkout hash", msg.CheckoutHash, w.Hash))
		}
		if w.InfoChan != wmh.infochan {
			t.Error("should be same channel")
		}
	}
	data := <- wmh.infochan
	if bytes.Compare(data, prBundleInfoMsg) != 0 {
		t.Error(util.StrFormatErrors("build data", "hit run pr bundle", string(data)))
	}
}

func TestWorkerMsgHandler_buildPushBuildBundle(t *testing.T) {
	msg := &protos.PushBuildBundle{
		CheckoutHash: "321321321321",
	}
	wmh := testGetWorkerMsgHandler(t, "testpushbuild")
	go wmh.build(msg)
	go func(){
		data := <- wmh.infochan
		if bytes.Compare(data, pushBundleInfoMsg) != 0 {
			t.Error(util.StrFormatErrors("build data", string(pushBundleInfoMsg), string(data)))
		}
	}()
}

func TestWorkerMsgHandler_UnmarshalAndProcess(t *testing.T) {
	wmh := testGetWorkerMsgHandler(t, "pr")
	message := &protos.PRBuildBundle{
		VaultToken:    "test vault token",
		CheckoutHash:  "test hash",
		PrData:        &protos.PullRequest{},
	}
	bytez, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	go wmh.UnmarshalAndProcess(bytez)
	go func(){
		data := <- wmh.infochan
		if bytes.Compare(data, prBundleInfoMsg) != 0 {
			t.Error(util.StrFormatErrors("build data", "hit run pr bundle", string(data)))
		}
	}()

	// unhappy path
	bytez = []byte("oh hey there, this is useless garbage data!! how nice!!")
	err = wmh.UnmarshalAndProcess(bytez)
	if err == nil {
		t.Error("should not pass unmarshaling")
	}
}