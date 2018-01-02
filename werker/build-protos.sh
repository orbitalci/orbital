#!/bin/sh
echo "building werker stream protobuf files"
protoc -I protobuf/ --go_out=plugins=grpc:protobuf/ buildoutputstream.proto
