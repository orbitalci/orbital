package commandhelper

import (
	"context"
	"crypto/tls"
	"flag"
	"net"
	"time"

	models "github.com/level11consulting/orbitalci/models/pb"
	//"github.com/level11consulting/orbitalci/util/secure_grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func GetClient(serverAddr string, noUseBlockingDial bool, tlsDns string) (client models.GuideOcelotClient, err error) {
	var conn *grpc.ClientConn
	if noUseBlockingDial {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		conn, err = grpc.Dial(serverAddr, opts...)
	} else {
		var tlsConf tls.Config
		creds := credentials.NewTLS(&tlsConf)
		ctx := context.Background()
		conn, err = BlockingDial(ctx, "tcp", serverAddr, creds)
	}
	client = models.NewGuideOcelotClient(conn)
	return
}

// BlockingDial is a helper method to dial the given address, using optional TLS credentials,
// and blocking until the returned connection is ready. If the given credentials are nil, the
// connection will be insecure (plain-text).
func BlockingDial(ctx context.Context, network, address string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// grpc.Dial doesn't provide any information on permanent connection errors (like
	// TLS handshake failures). So in order to provide good error messages, we need a
	// custom dialer that can provide that info. That means we manage the TLS handshake.
	result := make(chan interface{}, 1)

	writeResult := func(res interface{}) {
		// non-blocking write: we only need the first result
		select {
		case result <- res:
		default:
		}
	}

	dialer := func(address string, timeout time.Duration) (net.Conn, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		conn, err := (&net.Dialer{Cancel: ctx.Done()}).Dial(network, address)
		if err != nil {
			writeResult(err)
			return nil, err
		}
		if creds != nil {
			conn, _, err = creds.ClientHandshake(ctx, address, conn)
			if err != nil {
				writeResult(err)
				return nil, err
			}
		}
		return conn, nil
	}

	// Even with grpc.FailOnNonTempDialError, this call will usually timeout in
	// the face of TLS handshake errors. So we can't rely on grpc.WithBlock() to
	// know when we're done. So we run it in a goroutine and then use result
	// channel to either get the channel or fail-fast.
	go func() {
		opts = append(opts,
			grpc.WithBlock(),
			grpc.FailOnNonTempDialError(true),
			grpc.WithDialer(dialer),
			grpc.WithInsecure(), // we are handling TLS, so tell grpc not to
		)
		conn, err := grpc.DialContext(ctx, address, opts...)
		var res interface{}
		if err != nil {
			res = err
		} else {
			res = conn
		}
		writeResult(res)
	}()

	select {
	case res := <-result:
		if conn, ok := res.(*grpc.ClientConn); ok {
			return conn, nil
		}
		return nil, res.(error)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
