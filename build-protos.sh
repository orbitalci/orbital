#!/usr/bin/env sh
# ===============
cd protos 
./build-protos.sh

# inject our custom tags into build protobuf
protoc-go-inject-tag -input=./build.pb.go

cd ..
# ===============
cd werker
./build-protos.sh
cd ..
# ===============
# ===============
cd admin
./build-protos.sh
cd ..
# ===============

