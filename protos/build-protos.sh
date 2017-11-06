#!/usr/bin/env bash

# run protobuf compiler in root dir
cd ..

protoc --go_out=protos/out -I=protos protos/*.proto