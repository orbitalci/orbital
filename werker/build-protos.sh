#!/bin/sh
echo "building werker streamer protobuf files"
protoc -I protobuf/ --go_out=plugins=grpc:protobuf/ buildoutputstream.proto
