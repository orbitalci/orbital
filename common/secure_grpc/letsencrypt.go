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
		InsecureSkipVerify: true,
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
MIIGCzCCBPOgAwIBAgISBGKihHplX7yT6hNM3MmiluFuMA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0xODA0MTIxNTI4MjlaFw0x
ODA3MTExNTI4MjlaMBsxGTAXBgNVBAMTEG9jeWFkbWluLmwxMS5jb20wggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDfBmSEBV2cUmi9i3EKt9bOpPur0/uW
KqGvj/lutVhYrofKdIzQnjF1aZlTOm+MScR1L7ecTn+BS9bmTLI4FO0A/1LkO33d
02AlWFXkyRuemropBNsBh65epu3IH9RM8+c/v0EgZIjEh8i9u9OSnhP87W2928DU
TqzLFrajTf2anIDrrGa3uPiiq4RqjbB/2+E5QwQeSi8jyTXjB0GveW+bwiLD3dqE
ea98aJAVqrPAsMz/282mwFAsPQzFQBL+mP2cF7M25Pnf4T1ZA00kvGOHJAzLlKY5
TLcJjUq1d1td1jYfsj2j6mwJqG8ucRsSlBUv7NwBrZoQDy/Gsd2mcs69AgMBAAGj
ggMYMIIDFDAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFI61xZfrr7hgmzNbPT16elwY
gP4RMB8GA1UdIwQYMBaAFKhKamMEfd265tE5t6ZFZe/zqOyhMG8GCCsGAQUFBwEB
BGMwYTAuBggrBgEFBQcwAYYiaHR0cDovL29jc3AuaW50LXgzLmxldHNlbmNyeXB0
Lm9yZzAvBggrBgEFBQcwAoYjaHR0cDovL2NlcnQuaW50LXgzLmxldHNlbmNyeXB0
Lm9yZy8wGwYDVR0RBBQwEoIQb2N5YWRtaW4ubDExLmNvbTCB/gYDVR0gBIH2MIHz
MAgGBmeBDAECATCB5gYLKwYBBAGC3xMBAQEwgdYwJgYIKwYBBQUHAgEWGmh0dHA6
Ly9jcHMubGV0c2VuY3J5cHQub3JnMIGrBggrBgEFBQcCAjCBngyBm1RoaXMgQ2Vy
dGlmaWNhdGUgbWF5IG9ubHkgYmUgcmVsaWVkIHVwb24gYnkgUmVseWluZyBQYXJ0
aWVzIGFuZCBvbmx5IGluIGFjY29yZGFuY2Ugd2l0aCB0aGUgQ2VydGlmaWNhdGUg
UG9saWN5IGZvdW5kIGF0IGh0dHBzOi8vbGV0c2VuY3J5cHQub3JnL3JlcG9zaXRv
cnkvMIIBBAYKKwYBBAHWeQIEAgSB9QSB8gDwAHUAVYHUwhaQNgFK6gubVzxT8MDk
OHhwJQgXL6OqHQcT0wwAAAFiuq/aEgAABAMARjBEAiA2ART39LD+qkz+D2A8eZjU
ZY4DsxbjVkbw7M25NjNVHAIgRGWwVzFdn/josVQCX+1wY+olXZczSJKxi7bBRA3A
VoMAdwApPFGWVMg5ZbqqUPxYB9S3b79Yeily3KTDDPTlRUf0eAAAAWK6r9n7AAAE
AwBIMEYCIQDF8PbACRkmpyzBIqrawgL9A1jid5CjCxvTwrj2n9HXNgIhALWhicpP
WBdE95MII5hnhRpxBDu087PkNCUJYhc3aqFTMA0GCSqGSIb3DQEBCwUAA4IBAQB9
7m9N51yKL0XdLJoZMsABBjiD8ifYUhWH+f2ZuH7Oen36GEFyGTnjQ6jNoeMmc0/a
Gc5BfksIEkSb+Js5jdLsbTSnSM8OiZgSHT0re5qzTlckCBSRP/JjE1K5OgeMb6Ld
GoOFzhW4Ly8uEHkIDeyBOvkNiWAfp+m8Ztb6GRAcrsqvDKXvrvV0kU1aKSzcuFh7
fehpWvVsEP7CS1Ma4fnWkx9LVHG8xg28/8gNcakWbiukfaFDlJiJ4LFj/NCqY2n/
Tj2Gx3aDxbd+I2xooHx0Rbweo2mXOM3DGqNykZPlfugUsDDfE40gUc2et9hav//h
J03SkVoJS/OesgCB3MaE
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