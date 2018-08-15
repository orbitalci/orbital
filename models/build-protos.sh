#!/usr/bin/env sh

# run protobuf compiler on dir up
cd ..

echo "building bitbucket model proto files"
protoc --go_out=plugins=grpc:models/bitbucket/pb -I=models/bitbucket/ \
  -I$GOPATH/src \
  models/bitbucket/*.proto

echo "building slack model proto files"
protoc --go_out=models/slack/pb -I=models/slack/ \
  -I$GOPATH/src \
  models/slack/*.proto


echo "building root model proto files"

#uncomment if you want to generate javascript
#protoc --js_out=models/pb/js/ -I=models \
#  -I$GOPATH/src \
#  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
#  models/*.proto
protoc  --go_out=plugins=grpc:models/pb/ -I=models \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
  --grpc-gateway_out=logtostderr=true:models/pb \
  --swagger_out=logtostderr=true:models/pb \
  models/*.proto

#protoc --proto_path=src -I$GOPATH/src \
#  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
#  --js_out=library=protos.js,binary:pb src/foo.proto src/bar/baz.proto

echo "injecting custom tags"
# inject our custom tags into build protobuf
protoc-go-inject-tag -input=models/pb/guideocelot.pb.go
protoc-go-inject-tag -input=models/pb/build.pb.go
protoc-go-inject-tag -input=models/pb/creds.pb.go
protoc-go-inject-tag -input=models/bitbucket/pb/commonevententities.pb.go

# then we have to run go get in the stub directory cause grpc gateway ¯\_(ツ)_/¯ does this even work
cd models/pb
go get .