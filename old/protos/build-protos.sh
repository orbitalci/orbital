#!/usr/bin/env sh

echo "building top-level protobuf files"
# run protobuf compiler on dir up
cd ..

protoc --go_out=protos -I=protos protos/*.proto
# inject our custom tags into build protobuf
protoc-go-inject-tag -input=protos/build.pb.go