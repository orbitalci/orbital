package main

import (
	"errors"
	"github.com/shankj3/ocelot/util"
	"testing"
	"time"
)

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
	var testData = [][]byte{
		[]byte("ay ay ay"),
		[]byte("ze ze ze ze ze 27"),
		[]byte("1234a;slkdjf ze 27"),
		[]byte("ze ze ze ze 12 27"),
		[]byte("zequ queue queansdfa;lsdkjf garbage"),
	}
	var stream [][]byte
	ws := newWebSocketConn()
	//buildInfo.buildData = append()
	for _, dat := range testData {
		stream = append(stream, dat)
	}
	iterateOverBuildData(stream, ws)
	if !util.CompareByteArrays(stream, testData) {
		t.Errorf("arrays not the same. expected: %v, actual: %v", testData, stream)
	}
}

func Test_streamFromArray(t *testing.T) {
	// test data setup
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
	middleTest := testData[fstIndex:secIndex]
	middleActual := buildInfo.buildData[fstIndex:secIndex]
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