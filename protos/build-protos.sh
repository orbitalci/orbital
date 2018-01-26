#!/usr/bin/env sh

echo "building top-level protobuf files"
# run protobuf compiler on dir up
cd ..


protoc --go_out=protos -I=protos protos/*.proto
#protoc --go_out=Mpipeline.proto=bitbucket.org/level11consulting/leveler_resources:protos/ -I=protos/ -I=protos/leveler_resources/protos protos/*.proto
