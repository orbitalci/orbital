#!/usr/bin/env sh

echo "building top-level protobuf files"
# run protobuf compiler on dir up
cd ..


protoc --go_out=protos -I=protos protos/*.proto
