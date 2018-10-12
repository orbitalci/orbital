package bitbucket

//go:generate  protoc --go_out=plugins=grpc:pb/ -I . -I$GOPATH/src common.proto commonevententities.proto projectrootdir.proto webhook.proto respositories.proto
//go:generate protoc-go-inject-tag -input=pb/commonevententities.pb.go
