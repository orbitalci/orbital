package werker

import (
	"bytes"
	"testing"
	"bitbucket.org/level11consulting/go-til/test"
	pb "github.com/shankj3/ocelot/protos"
	"encoding/gob"

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
			//t.Log("recieved")
			if bytes.Compare(info, wd.chanData) != 0 {
				t.Error(test.StrFormatErrors("info channel response", string(wd.chanData), string(info)))
			}
			if wd.hash != trans.Hash {
				t.Error(test.StrFormatErrors("git hash", wd.hash, trans.Hash))
			}
		})
	}
}

func TestWorkerMsgHandler_buildPRBuildBundle(t *testing.T) {
	msg := &pb.WerkerTask{
		CheckoutHash: "1231231231234",
	}
	wmh := testGetWorkerMsgHandler(t, "testbuild")
	go wmh.build(msg)
	w, ok := <-wmh.ChanChan; if !ok {
		t.Fatal("no data on channel")
	} else {
		if w.Hash != msg.CheckoutHash {
			t.Error(test.StrFormatErrors("checkout hash", msg.CheckoutHash, w.Hash))
		}
		if w.InfoChan != wmh.infochan {
			t.Error("should be same channel")
		}
	}
	data := <- wmh.infochan
	if bytes.Compare(data, prBundleInfoMsg) != 0 {
		t.Error(test.StrFormatErrors("build data", "hit run pr bundle", string(data)))
	}
}

func TestWorkerMsgHandler_buildPushBuildBundle(t *testing.T) {
	msg := &pb.WerkerTask{
		CheckoutHash: "321321321321",
	}
	wmh := testGetWorkerMsgHandler(t, "testpushbuild")
	go wmh.build(msg)
	w, ok := <-wmh.ChanChan; if !ok {
		t.Fatal("no data on channel")
	} else {
		if w.Hash != msg.CheckoutHash {
			t.Error(test.StrFormatErrors("checkout hash", msg.CheckoutHash, w.Hash))
		}
		if w.InfoChan != wmh.infochan {
			t.Error("should be same channel")
		}
	}
	data := <- wmh.infochan
	if bytes.Compare(data, pushBundleInfoMsg) != 0 {
		t.Error(test.StrFormatErrors("build data", string(pushBundleInfoMsg), string(data)))
	}
}

func TestMyTheory(t *testing.T) {
	message := pb.WerkerTask {
		VaultToken: "test vault token",
		CheckoutHash:  "test hash",
	}

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	gob.Register(message)
	err := enc.Encode(message)

	if err != nil {
		t.Error(err)
	}

	werkerBytes := buf.Bytes()
	var decoderBuf bytes.Buffer

	dec := gob.NewDecoder(&decoderBuf)
	err = dec.Decode(&werkerBytes)
	if err != nil {
		t.Error(err)
	}

	if len(decoderBuf.Bytes()) == 0 {
		t.Error("dear god you suck")
	}
}

func TestWorkerMsgHandler_UnmarshalAndProcess(t *testing.T) {
	wmh := testGetWorkerMsgHandler(t, "build")
	message := pb.WerkerTask {
		VaultToken: "test vault token",
		CheckoutHash:  "test hash",
	}

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	gob.Register(message)
	err := enc.Encode(message)

	if err != nil {
		t.Fatal(err)
	}
	go wmh.UnmarshalAndProcess(buf.Bytes())
	w, ok := <-wmh.ChanChan; if !ok {
		t.Fatal("no data on channel")
	} else {
		if w.Hash != message.CheckoutHash {
			t.Error(test.StrFormatErrors("checkout hash", message.CheckoutHash, w.Hash))
		}
	}
	data := <- w.InfoChan
	if bytes.Compare(data, prBundleInfoMsg) != 0 {
		t.Error(test.StrFormatErrors("build data", "hit run pr bundle", string(data)))
	}


	// unhappy path
	bytez := []byte("oh hey there, this is useless garbage data!! how nice!!")
	err = wmh.UnmarshalAndProcess(bytez)
	if err == nil {
		t.Error("should not pass unmarshaling")
	}
}