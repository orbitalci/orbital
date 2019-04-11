package streaminglogs

import (
	"testing"
	"time"

	"github.com/go-test/deep"
	protobuf "github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"google.golang.org/grpc"
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

func (t *testBuildInfoGrpcServer) SendIt(data []byte) error {
	err := t.Send(&protobuf.Response{OutputLine: string(data)})
	return err
}

func (t *testBuildInfoGrpcServer) SendError(errorDes []byte) {}

func (t *testBuildInfoGrpcServer) Finish(chan int) {}

func Test_iterateOverBuildData(t *testing.T) {
	var stream = NewTestStreamArray()
	ws := ocenet.NewWebSocketConn()
	//buildInfo.buildData = append()
	for _, dat := range testData {
		stream.data = append(stream.data, dat)
	}
	_, err := iterateOverByteArray(stream, ws, 0)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(ws.MsgData, testData); diff != nil {
		t.Error(diff)
	}
	var streamGrpc = NewTestStreamArray()
	grp := &testBuildInfoGrpcServer{}
	for _, datum := range testData {
		streamGrpc.data = append(streamGrpc.data, datum)
	}
	_, err = iterateOverByteArray(streamGrpc, grp, 0)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(grp.testData, stringTestData); diff != nil {
		t.Error(diff)
	}
}

func Test_streamFromArray(t *testing.T) {
	// test data setup
	var fstIndex = 4
	var secIndex = 6
	var streamobj = NewTestStreamArray()
	var ws = ocenet.NewWebSocketConn()
	//
	streamobj.AddToData(testData[:fstIndex])
	go StreamFromArray(streamobj, ws, log.Log())
	time.Sleep(1 * time.Second)
	if diff := deep.Equal(testData[:fstIndex], ws.MsgData); diff != nil {
		t.Error(diff)
	}
	if diff := deep.Equal(testData[:fstIndex], streamobj.GetData()); diff != nil {
		t.Error(diff)
	}
	streamobj.AddToData(testData[fstIndex:secIndex])

	time.Sleep(1 * time.Second)
	middleTest := testData[:secIndex]
	middleActual := ws.MsgData[:secIndex]
	if diff := deep.Equal(middleTest, middleActual); diff != nil {
		t.Error(diff)
	}
	if diff := deep.Equal(middleTest, streamobj.GetData()[:secIndex]); diff != nil {
		t.Error(diff)
	}

	streamobj.AddToData(testData[secIndex:])
	time.Sleep(1 * time.Second)
	if diff := deep.Equal(testData, ws.MsgData); diff != nil {
		t.Error(diff)
	}
}
