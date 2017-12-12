package secure_grpc

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/philips/grpc-gateway-example/insecure"
	"google.golang.org/grpc/credentials"
)

type fakeSecure struct {
	fakeCert *x509.CertPool
}

func NewFakeSecure() *fakeSecure {
	fc := x509.NewCertPool()
	ok := fc.AppendCertsFromPEM([]byte(insecure.Cert))
	if !ok {
		panic("bad certs")
	}
	return &fakeSecure{
		fakeCert: fc,
	}
}

// GetNewClientTLS will create a fake cert using x509 and the insecure package. it will be enough
// to have tls but generate a warning. I'm pulling this out because the Client will need it too.
func (fs *fakeSecure) GetNewClientTLS(serverRunsAt string) credentials.TransportCredentials {
	ok := fs.fakeCert.AppendCertsFromPEM([]byte(insecure.Cert))
	if !ok {
		panic("bad certs")
	}
	return credentials.NewClientTLSFromCert(fs.fakeCert, serverRunsAt)
}

func (fs *fakeSecure) GetNewTLS(serverRunsAt string) credentials.TransportCredentials {
	return credentials.NewTLS(&tls.Config{
		ServerName: serverRunsAt,
		RootCAs:    fs.fakeCert,
	})
}


func (fs *fakeSecure) GetKeyPair() *tls.Certificate {
	pair, err := tls.X509KeyPair([]byte(insecure.Cert), []byte(insecure.Key))
	if err != nil {
		panic("couldn't get tls x509 key pair with insecure certs")
	}
	return &pair
}