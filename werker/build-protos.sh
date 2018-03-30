#!/usr/bin/env sh
echo "building werker stream protobuf files"

protoc -I protobuf/ --go_out=plugins=grpc:protobuf/ protobuf/werkerserver.proto
