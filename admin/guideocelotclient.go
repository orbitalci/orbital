package admin

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/secure_grpc"
	"flag"
	"google.golang.org/grpc"
)

var (
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)


func GetClient(serverAddr string, insecure bool, tlsDns string) (client models.GuideOcelotClient, err error){
	var secure secure_grpc.SecureGrpc
	if insecure {
		secure = secure_grpc.NewFakeSecure()
	} else {
		secure = secure_grpc.NewLeSecure()
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(secure.GetNewClientTLS(tlsDns)))
	conn, err := grpc.Dial(serverAddr, opts...)
	client = models.NewGuideOcelotClient(conn)
	return
}