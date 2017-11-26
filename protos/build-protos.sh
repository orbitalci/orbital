#!/usr/bin/env bash

# run protobuf compiler on dir up
cd ..

protoc --go_out=protos/out -I=protos protos/*.proto