package main

import (
	"bufio"
	"errors"
	"github.com/shankj3/ocelot/util"
	"github.com/shankj3/ocelot/util/consulet"
	"github.com/shankj3/ocelot/util/storage"
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

type testWebSocketConn struct {
	msgData [][]byte
	done    bool
}

func (ws *testWebSocketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (ws *testWebSocketConn) WriteMessage(messageType int, data []byte) error {
	if !ws.done {
		ws.msgData = append(ws.msgData, data)
		return nil
	} else {
		return errors.New("can't write to done msg")
	}
}

func (ws *testWebSocketConn) Close() error {
	ws.done = true
	return nil
}

func (ws *testWebSocketConn) GetData() [][]byte {
	return ws.msgData
}

func newWebSocketConn() *testWebSocketConn {
	var data [][]byte
	return &testWebSocketConn{
		msgData: data,
		done:    false,
	}
}

func Test_iterateOverBuildData(t *testing.T) {
	var stream [][]byte
	ws := newWebSocketConn()
	//buildInfo.buildData = append()
	for _, dat := range testData {
		stream = append(stream, dat)
	}
	iterateOverBuildData(stream, ws)
	if !util.CompareByteArrays(ws.msgData, testData) {
		t.Errorf("arrays not the same. expected: %v, actual: %v", testData, ws.msgData)
	}
}

func Test_streamFromArray(t *testing.T) {
	// test data setup
	var stream [][]byte
	var fstIndex  = 4
	var secIndex  = 6
	var buildInfo = &buildDatum{buildData: stream, done: make(chan int),}
	var ws        = newWebSocketConn()
	//
	for _, data := range testData[:fstIndex] {
		buildInfo.buildData = append(buildInfo.buildData, data)
	}
	go streamFromArray(buildInfo, ws)
	time.Sleep(1 * time.Second)
	if !util.CompareByteArrays(testData[:fstIndex], ws.msgData) {
		t.Errorf("first slices not the same. expected: %v, actual: %v", testData[:fstIndex],  buildInfo.buildData)
	}
	for _, data := range testData[fstIndex:secIndex] {
		buildInfo.buildData = append(buildInfo.buildData, data)
	}
	time.Sleep(1 * time.Second)
	middleTest := testData[:secIndex]
	middleActual := buildInfo.buildData[:secIndex]
	if !util.CompareByteArrays(middleTest, middleActual) {
		t.Errorf("second slices not the same. expected: %v, actual: %v", testData, stream)
	}
	for _, data := range testData[secIndex:] {
		buildInfo.buildData = append(buildInfo.buildData, data)
	}
	if !util.CompareByteArrays(testData, buildInfo.buildData) {
		t.Errorf("full arrays not the same. expected: %v, actual: %v", testData, stream)
	}
}

func Test_writeInfoChanToInMemMap(t *testing.T) {
	trans := &Transport{"FOR_TESTING", make(chan []byte)}
	ctx := &appContext{
		buildInfo: make(map[string]*buildDatum),
		storage: storage.NewFileBuildStorage(""),
		consul: consulet.Default(),
	}
	middleIndex := 6
	go writeInfoChanToInMemMap(trans, ctx)
	for _, data := range testData[:middleIndex] {
		trans.InfoChan <- data
	}
	time.Sleep(100)
	if !util.CompareByteArrays(testData[:middleIndex], ctx.buildInfo[trans.Hash].buildData) {
		t.Errorf("middle slice not the same. expected: %v, actual: %v", testData[:middleIndex], ctx.buildInfo[trans.Hash].buildData)
	}
	for _, data := range testData[middleIndex:] {
		trans.InfoChan <- data
	}
	time.Sleep(100)
	if !util.CompareByteArrays(testData, ctx.buildInfo[trans.Hash].buildData) {
		t.Errorf("full slice not the same. expected: %v, actual: %v", testData, ctx.buildInfo[trans.Hash].buildData)
	}
	close(trans.InfoChan)
	time.Sleep(100)

	// wait for storage to be done, then check it
	<- ctx.buildInfo[trans.Hash].done
	reader, err := ctx.storage.RetrieveReader(trans.Hash)
	if err != nil {
		t.Fatal(err)
	}
	var actualData [][]byte
	// todo: this is a dumb and lazy and nonperformant way but its late
	sc := bufio.NewScanner(reader)
	for sc.Scan() {
		actualData = append(actualData, sc.Bytes())
	}
	if !util.CompareByteArrays(testData, actualData) {
		t.Errorf("bytes from storage not same as testdata. expected: %v, actual: %v", testData, actualData)
	}
	// todo: check the consul stuff
}