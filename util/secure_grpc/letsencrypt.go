package secure_grpc

import (
	"crypto/tls"
	"crypto/x509"
	//"github.com/philips/grpc-gateway-example/insecure"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

type leSecure struct {
	cert *x509.CertPool
}

func NewLeSecure() *leSecure {
	fc := x509.NewCertPool()
	ok := fc.AppendCertsFromPEM([]byte(fullChain))
	if !ok {
		panic("bad certs")
	}
	return &leSecure{
		cert: fc,
	}
}

// GetNewClientTLS will create a fake cert using x509 and the insecure package. it will be enough
// to have tls but generate a warning. I'm pulling this out because the Client will need it too.
func (fs *leSecure) GetNewClientTLS(serverRunsAt string) credentials.TransportCredentials {
	ok := fs.cert.AppendCertsFromPEM([]byte(fullChain))
	if !ok {
		panic("bad certs")
	}
	return credentials.NewClientTLSFromCert(fs.cert, serverRunsAt)
}

func (fs *leSecure) GetNewTLS(serverRunsAt string) credentials.TransportCredentials {
	return credentials.NewTLS(&tls.Config{
		ServerName: serverRunsAt,
		RootCAs:    fs.cert,
	})
}


func (fs *leSecure) GetKeyPair() *tls.Certificate {
	key, err := ioutil.ReadFile("/etc/letsencrypt/live/ocelot.hq.l11.com/privkey.pem")
	if err != nil {
		panic("couldn't get private certs" + err.Error())
	}
	pair, err := tls.X509KeyPair([]byte(fullChain), key)
	if err != nil {
		panic("couldn't get tls x509 key pair with insecure certs")
	}
	return &pair
}


var fullChain = `-----BEGIN CERTIFICATE-----
MIIFBTCCA+2gAwIBAgISA4GA8jVATdolNKmdCeFuisBxMA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0xNzEyMjMxOTAwMzJaFw0x
ODAzMjMxOTAwMzJaMBwxGjAYBgNVBAMTEW9jZWxvdC5ocS5sMTEuY29tMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArnkFa/6PQzTQEl86Ml4yEbGUR9uD
TS2nwnZFxd+0I+LDGL721uT1Rl/x8wPFI1Fojl3DtSvG8TVQWhupgFpVCy/ayL7p
pJdV/+qoXnYO74NO6TFdCqV70lXcN6P+HsAWyBis7dWMPQCscZEADtSb6hz3pYQg
ragTISLBbggY/yOKl/GyYjuZ7XSTaJOUg9UJYNr4LnkNhLmdUSsZAvJkh2cRRzcx
Se2C+oxLpzHXhgvfWM/jx4FUIHQ+Qzilgs2My9PziNdC7sbMoxeBHsi9LNBJQR8+
FVsjgNeq9k6F9U7kJs0O8E+aivq2MKy2JHtB9iu4PEPHN6DB83cEFqWQxQIDAQAB
o4ICETCCAg0wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggr
BgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBS6zsyQjdqGMMrqQ52IjkAZ
H8Y/ejAfBgNVHSMEGDAWgBSoSmpjBH3duubRObemRWXv86jsoTBvBggrBgEFBQcB
AQRjMGEwLgYIKwYBBQUHMAGGImh0dHA6Ly9vY3NwLmludC14My5sZXRzZW5jcnlw
dC5vcmcwLwYIKwYBBQUHMAKGI2h0dHA6Ly9jZXJ0LmludC14My5sZXRzZW5jcnlw
dC5vcmcvMBwGA1UdEQQVMBOCEW9jZWxvdC5ocS5sMTEuY29tMIH+BgNVHSAEgfYw
gfMwCAYGZ4EMAQIBMIHmBgsrBgEEAYLfEwEBATCB1jAmBggrBgEFBQcCARYaaHR0
cDovL2Nwcy5sZXRzZW5jcnlwdC5vcmcwgasGCCsGAQUFBwICMIGeDIGbVGhpcyBD
ZXJ0aWZpY2F0ZSBtYXkgb25seSBiZSByZWxpZWQgdXBvbiBieSBSZWx5aW5nIFBh
cnRpZXMgYW5kIG9ubHkgaW4gYWNjb3JkYW5jZSB3aXRoIHRoZSBDZXJ0aWZpY2F0
ZSBQb2xpY3kgZm91bmQgYXQgaHR0cHM6Ly9sZXRzZW5jcnlwdC5vcmcvcmVwb3Np
dG9yeS8wDQYJKoZIhvcNAQELBQADggEBAHHZj43xcRlrR66KfwibZ/HIFGxpprTl
CrtzqtkuQAWF1DCDwEzeaHZMq5+s1w4Py7iJHeEOaLeV2tM3EdQkb8gmqXipQ6k5
UFTTTfPBTF+awcwAR5icKXcUudPV+sqUgAikuLrXPvHekS5RhPxn/6NWBDCg3Dab
cDYAqiYmfc+10YlkC3+WmAHBzyYuNzparbIrlwMdC4Wy1AGVfQaL3EqGt1ssQHu4
G+fHiINmfnGV+YO+Ci9cyXaUEsTg2rRDlUPwCEe4nyU2MlJUwbcEHcPmMnKZ6tco
+jgYtCvqPnnBSw8b4ZWRURxD4nCtA0GoIuXsNCJV6bLJIUdQ+xcziSI=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEkjCCA3qgAwIBAgIQCgFBQgAAAVOFc2oLheynCDANBgkqhkiG9w0BAQsFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTE2MDMxNzE2NDA0NloXDTIxMDMxNzE2NDA0Nlow
SjELMAkGA1UEBhMCVVMxFjAUBgNVBAoTDUxldCdzIEVuY3J5cHQxIzAhBgNVBAMT
GkxldCdzIEVuY3J5cHQgQXV0aG9yaXR5IFgzMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAnNMM8FrlLke3cl03g7NoYzDq1zUmGSXhvb418XCSL7e4S0EF
q6meNQhY7LEqxGiHC6PjdeTm86dicbp5gWAf15Gan/PQeGdxyGkOlZHP/uaZ6WA8
SMx+yk13EiSdRxta67nsHjcAHJyse6cF6s5K671B5TaYucv9bTyWaN8jKkKQDIZ0
Z8h/pZq4UmEUEz9l6YKHy9v6Dlb2honzhT+Xhq+w3Brvaw2VFn3EK6BlspkENnWA
a6xK8xuQSXgvopZPKiAlKQTGdMDQMc2PMTiVFrqoM7hD8bEfwzB/onkxEz0tNvjj
/PIzark5McWvxI0NHWQWM6r6hCm21AvA2H3DkwIDAQABo4IBfTCCAXkwEgYDVR0T
AQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAYYwfwYIKwYBBQUHAQEEczBxMDIG
CCsGAQUFBzABhiZodHRwOi8vaXNyZy50cnVzdGlkLm9jc3AuaWRlbnRydXN0LmNv
bTA7BggrBgEFBQcwAoYvaHR0cDovL2FwcHMuaWRlbnRydXN0LmNvbS9yb290cy9k
c3Ryb290Y2F4My5wN2MwHwYDVR0jBBgwFoAUxKexpHsscfrb4UuQdf/EFWCFiRAw
VAYDVR0gBE0wSzAIBgZngQwBAgEwPwYLKwYBBAGC3xMBAQEwMDAuBggrBgEFBQcC
ARYiaHR0cDovL2Nwcy5yb290LXgxLmxldHNlbmNyeXB0Lm9yZzA8BgNVHR8ENTAz
MDGgL6AthitodHRwOi8vY3JsLmlkZW50cnVzdC5jb20vRFNUUk9PVENBWDNDUkwu
Y3JsMB0GA1UdDgQWBBSoSmpjBH3duubRObemRWXv86jsoTANBgkqhkiG9w0BAQsF
AAOCAQEA3TPXEfNjWDjdGBX7CVW+dla5cEilaUcne8IkCJLxWh9KEik3JHRRHGJo
uM2VcGfl96S8TihRzZvoroed6ti6WqEBmtzw3Wodatg+VyOeph4EYpr/1wXKtx8/
wApIvJSwtmVi4MFU5aMqrSDE6ea73Mj2tcMyo5jMd6jmeWUHK8so/joWUoHOUgwu
X4Po1QYz+3dszkDqMp4fklxBwXRsW10KXzPMTZ+sOPAveyxindmjkW8lGy+QsRlG
PfZ+G6Z6h7mjem0Y+iWlkYcV4PIWL1iwBi8saCbGS5jN2p8M+X+Q7UNKEkROb3N6
KOqkqm57TH2H3eDJAkSnh6/DNFu0Qg==
-----END CERTIFICATE-----`