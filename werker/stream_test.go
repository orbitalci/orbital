package werker

import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"bufio"
	"bytes"
	"google.golang.org/grpc"
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

func stringifyTestData(testData [][]byte) []string {
	var stringy []string
	for _, line := range testData {
		stringy = append(stringy, string(line))
	}
	return stringy
}

var stringTestData = stringifyTestData(testData)

type testBuildInfoGrpcServer struct {
	testData []string
	grpc.ServerStream
}

func (t *testBuildInfoGrpcServer) Send(response *protobuf.Response) error {
	t.testData = append(t.testData, response.OutputLine)
	return nil
}

func Test_iterateOverBuildData(t *testing.T) {
	var stream [][]byte
	ws := ocenet.NewWebSocketConn()
	//buildInfo.buildData = append()
	for _, dat := range testData {
		stream = append(stream, dat)
	}
	iterateOverBuildData(stream, ws)
	if !test.CompareByteArrays(ws.MsgData, testData) {
		t.Errorf("arrays not the same. expected: %v, actual: %v", testData, ws.MsgData)
	}
	var streamGrpc [][]byte
	grp := &testBuildInfoGrpcServer{}
	for _, datum := range testData {
		streamGrpc = append(streamGrpc, datum)
	}
	iterateOverBuildData(streamGrpc, grp)
	if !test.CompareStringArrays(grp.testData, stringTestData) {
		t.Errorf("arrays not same for grpc. expected: %s, actual: %s", stringTestData, grp.testData)
	}
}

func Test_streamFromArray(t *testing.T) {
	// test data setup
	var stream [][]byte
	var fstIndex = 4
	var secIndex = 6
	var buildInfo = &buildDatum{buildData: stream, done: false}
	var ws = ocenet.NewWebSocketConn()
	//
	for _, data := range testData[:fstIndex] {
		buildInfo.buildData = append(buildInfo.buildData, data)
	}
	go streamFromArray(buildInfo, ws)
	time.Sleep(1 * time.Second)
	if !test.CompareByteArrays(testData[:fstIndex], ws.MsgData) {
		t.Errorf("first slices not the same. expected: %v, actual: %v", testData[:fstIndex], buildInfo.buildData)
	}
	for _, data := range testData[fstIndex:secIndex] {
		buildInfo.buildData = append(buildInfo.buildData, data)
	}
	time.Sleep(1 * time.Second)
	middleTest := testData[:secIndex]
	middleActual := buildInfo.buildData[:secIndex]
	if !test.CompareByteArrays(middleTest, middleActual) {
		t.Errorf("second slices not the same. expected: %v, actual: %v", testData, buildInfo.buildData[:secIndex])
	}
	for _, data := range testData[secIndex:] {
		buildInfo.buildData = append(buildInfo.buildData, data)
	}
	if !test.CompareByteArrays(testData, buildInfo.buildData) {
		t.Errorf("full arrays not the same. expected: %v, actual: %v", testData, buildInfo.buildData)
	}
}

func Test_writeInfoChanToInMemMap(t *testing.T) {
	trans := &Transport{"FOR_TESTING", make(chan []byte)}
	werkerConsulet, _ := consulet.Default()
	ctx := &werkerStreamer{
		buildInfo: make(map[string]*buildDatum),
		storage:   storage.NewFileBuildStorage(""),
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

	// wait for storage to be done, then check it
	for !ctx.buildInfo[trans.Hash].done {
		time.Sleep(100)
	}
	bytez, err := ctx.storage.Retrieve(trans.Hash)
	if err != nil {
		t.Fatal(err)
	}
	reader := bytes.NewReader(bytez)
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
	fbs := ctx.storage.(*storage.FileBuildStorage)
	defer fbs.Clean()
	// todo: check the consul stuff
}
