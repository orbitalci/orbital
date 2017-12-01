#!/bin/bash

protoc -I models/ -I. \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --go_out=plugins=grpc:models \
    models/guideocelot.proto

# generate reverse proxy cause grpc gateway
protoc -I models/ -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=logtostderr=true:models \
  models/guideocelot.proto