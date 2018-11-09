package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/secure_grpc"
	"github.com/shankj3/ocelot/storage"
)

func Test_prep(t *testing.T) {
	_, _, _, _, err := GetGrpcServer(&rc{}, secure_grpc.NewFakeSecure(), "localhost", "9199", "9198", "")
	if err != nil {
		t.Error(err)
	}
	//cleanup()
	//serv.GracefulStop()
	_, _, _, _, err = GetGrpcServer(&rc{nostore: true}, secure_grpc.NewFakeSecure(), "localhost", "9099", "9098", "")
	if err == nil {
		t.Error("should fail, as did not return storage")
	}
	//cleanup()
	//serv.GracefulStop()
}

func Test_Start(t *testing.T) {
	serv, listen, _, _, err := GetGrpcServer(&rc{}, secure_grpc.NewFakeSecure(), "localhost", "9999", "9998", "")
	if err != nil {
		t.Error(err)
	}
	go Start(serv, listen)
	serv.GracefulStop()
}

func Test_serveSwagger(t *testing.T) {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/swag/swag.swagger.json", nil)
	serveSwagger(resp, req)
	if resp.Code != http.StatusNotFound {
		t.Error("swag.swagger.json not a real file")
	}
	req = httptest.NewRequest("GET", "/swag/swag", nil)
	serveSwagger(resp, req)
	if resp.Code != http.StatusNotFound {
		t.Error("doesn't end with .swagger.json, should return 404")
	}
}

func Test_preflightHandler(t *testing.T) {
	resp := httptest.NewRecorder()
	preflightHandler(resp, httptest.NewRequest("GET", "/.well-known/acme-challenge/token-http-01", nil))
	methods := resp.Header().Get("Access-Control-Allow-Methods")
	if methods != "GET,HEAD,POST,PUT,DELETE" {
		t.Error("wrong acces controll allow methods, got " + methods)
	}
	headers := resp.Header().Get("Access-Control-Allow-Headers")
	if headers != "Content-Type,Accept" {
		t.Error("wrongg header, got " + headers)
	}
}

type rc struct {
	nostore bool
	credentials.CVRemoteConfig
}

func (r *rc) GetOcelotStorage() (storage.OcelotStorage, error) {
	if r.nostore {
		return nil, errors.New("nostore4u")
	}

	return &st{}, nil
}

type st struct {
	storage.OcelotStorage
}

func (s *st) Close() {}
