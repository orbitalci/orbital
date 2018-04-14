package streamer

import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bufio"
	"bytes"
	"testing"
	"time"
)

var testData = [][]byte{
	[]byte("ay ay ay"),
	[]byte("ze ze ze ze ze 27"),
	[]byte("1234a;slkdjf ze 27"),
	[]byte("ze ze ze ze 12 27"),
	[]byte("zequ queue queansdfa;lsdkjf garbage"),
	[]byte("zequ queue queanoai3rnfe"),
	[]byte("zequ que12321age"),
	[]byte("z-3985jfap0s9en;dopu5;lsdkjf garbage"),
	[]byte("zequ qu123eue queanoai3rnfe"),
	[]byte("zequ que12457fd321age"),
	[]byte("z-3985jfasr jhgfturkeyisgrossf garbage"),
}

func Test_writeInfoChanToInMemMap(t *testing.T) {
	store := storage.NewFileBuildStorage("./test-fixtures/store")
	id, err := store.AddSumStart("FOR_TESTING", "myacct", "myrepo", "BRANCH!")
	if err != nil {
		t.Fatal("could not create setup data, err: ", err.Error())
	}
	trans := &Transport{Hash: "FOR_TESTING", InfoChan: make(chan []byte), DbId: id}
	werkerConsulet, _ := consulet.Default()
	ctx := &WerkerContext{
		buildInfo: make(map[string]*buildDatum),
		out: 	    store,
		consul:    werkerConsulet,
	}
	middleIndex := 6
	go writeInfoChanToInMemMap(trans, ctx)
	for _, data := range testData[:middleIndex] {
		trans.InfoChan <- data
	}
	time.Sleep(100)
	if !test.CompareByteArrays(testData[:middleIndex], ctx.buildInfo[trans.Hash].buildData) {
		t.Errorf("middle slice not the same. expected: %v, actual: %v", testData[:middleIndex], ctx.buildInfo[trans.Hash].buildData)
	}
	for _, data := range testData[middleIndex:] {
		trans.InfoChan <- data
	}
	time.Sleep(100)
	if !test.CompareByteArrays(testData, ctx.buildInfo[trans.Hash].buildData) {
		t.Errorf("full slice not the same. expected: %v, actual: %v", testData, ctx.buildInfo[trans.Hash].buildData)
	}
	close(trans.InfoChan)

	// wait for out to be done, then check it
	for !ctx.buildInfo[trans.Hash].done {
		time.Sleep(100)
	}
	out, err := ctx.out.RetrieveLastOutByHash(trans.Hash)
	if err != nil {
		t.Fatal(err)
	}
	reader := bytes.NewReader(out.Output)
	var actualData [][]byte
	// todo: this is a dumb and lazy and nonperformant way but its late
	sc := bufio.NewScanner(reader)
	for sc.Scan() {
		actualData = append(actualData, sc.Bytes())
	}
	if !test.CompareByteArrays(testData, actualData) {
		t.Errorf("bytes from storage not same as testdata. expected: %v, actual: %v", testData, actualData)
	}
	// remove stored test data
	fbs := ctx.out.(*storage.FileBuildStorage)
	defer fbs.Clean()
	// todo: check the consul stuff
}
