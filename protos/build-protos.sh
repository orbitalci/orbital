#!/usr/bin/env sh

# run protobuf compiler on dir up
cd ..

#protoc --go_out=protos -I=protos/leveler_resources/protos protos/leveler_resources/protos/*.proto
protoc --go_out=Mpipeline.proto=github.com/shankj3/ocelot/protos/leveler_resources:protos/ -I=protos/ -I=protos/leveler_resources/protos protos/*.proto
#protoc --go_out=protos -I=protos -I=protos/leveler_resources/protos protos/*.proto
