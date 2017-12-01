#!/usr/bin/env sh

# run protobuf compiler on dir up
cd ..

protoc --go_out=protos/out -I=protos protos/*.proto
