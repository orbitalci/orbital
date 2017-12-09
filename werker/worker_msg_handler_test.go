package werker

import (
	"bytes"
	"testing"
	"bitbucket.org/level11consulting/go-til/test"
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