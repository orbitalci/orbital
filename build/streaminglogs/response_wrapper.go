package streaminglogs

import (
	"github.com/level11consulting/ocelot/models/pb"
)

//RespWrap will wrap streaming messages in a LineResponse object to be sent by the server stream
func RespWrap(msg string) *pb.LineResponse {
	return &pb.LineResponse{OutputLine: msg}
}
