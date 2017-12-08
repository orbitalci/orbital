#!/usr/bin/env sh

# run protobuf compiler on dir up
cd ..

#protoc --go_out=protos -I=protos/leveler_resources/protos protos/leveler_resources/protos/*.proto
protoc --go_out=protos -I=protos -I=Mprotos/leveler_resources/protos=bitbucket.org/level11consulting/leveler_resources/protos protos/*.proto
