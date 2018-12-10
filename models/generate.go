package models

//go:generate protoc --go_out=plugins=grpc:pb/ -I=. -I$GOPATH/src -I${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.6.2/third_party/googleapis/ -I ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.6.2 --grpc-gateway_out=logtostderr=true:pb --swagger_out=logtostderr=true:pb guideocelot.proto storage.proto build.proto creds.proto vcshandler.proto werkerserver.proto

//go:generate protoc-go-inject-tag -input=pb/guideocelot.pb.go
//go:generate protoc-go-inject-tag -input=pb/build.pb.go
//go:generate protoc-go-inject-tag -input=pb/creds.pb.go
