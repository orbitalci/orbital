package models


//go:generate protoc --go_out=plugins=grpc:pb/ -I=. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway --grpc-gateway_out=logtostderr=true:pb --swagger_out=logtostderr=true:pb guideocelot.proto storage.proto build.proto creds.proto vcshandler.proto werkerserver.proto

//go:generate protoc-go-inject-tag -input=pb/guideocelot.pb.go
//go:generate protoc-go-inject-tag -input=pb/build.pb.go
//go:generate protoc-go-inject-tag -input=pb/creds.pb.go
