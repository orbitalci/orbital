#!/usr/bin/env bash

# run protobuf compiler in root dir
cd ..

protoc --go_out=protos -I=protos protos/*.proto