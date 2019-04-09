package secure_grpc

import (
	"crypto/tls"
	"google.golang.org/grpc/credentials"
)

type SecureGrpc interface {
	GetNewClientTLS(serverRunsAt string) credentials.TransportCredentials
	GetNewTLS(serverRunsAt string) credentials.TransportCredentials
	GetKeyPair() *tls.Certificate
}
