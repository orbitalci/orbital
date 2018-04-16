#!/usr/bin/env sh

echo "building admin protobuf files"

echo "[DEBUG] first"
protoc -I models/ -I. \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
    --go_out=plugins=grpc:models \
    models/guideocelot.proto

echo "[DEBUG] second"
# generate reverse proxy cause grpc gateway
protoc -I models/ -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
  --grpc-gateway_out=logtostderr=true:models \
  --swagger_out=logtostderr=true:models \
  models/guideocelot.proto

# inject our custom tags into build protobuf
protoc-go-inject-tag -input=models/guideocelot.pb.go

echo "[DEBUG] third"
# then we have to run go get in the stub directory cause grpc gateway ¯\_(ツ)_/¯ does this even work
cd models
go get .
