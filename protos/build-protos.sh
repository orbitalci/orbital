#!/usr/bin/env sh

# run protobuf compiler on dir up
cd ..

protoc --go_out=protos/out -I=/home/mariannefeng/go/src/leveler/protobuf -I=protos protos/*.proto
