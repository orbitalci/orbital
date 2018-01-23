#!/usr/bin/env sh

echo "building admin protobuf files"

echo "[DEBUG] first"
protoc -I models/ -I. \
    -I/usr/local/include \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --go_out=plugins=grpc:models \
    models/guideocelot.proto

echo "[DEBUG] second"
# generate reverse proxy cause grpc gateway
protoc -I models/ -I. \
  -I/usr/local/include \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=logtostderr=true:models \
  models/guideocelot.proto

echo "[DEBUG] third"
# then we have to run go get in the stub directory cause grpc gateway ¯\_(ツ)_/¯ does this even work
cd models
go get .
