package github

//go:generate  protoc --go_out=plugins=grpc:pb/ -I . -I$GOPATH/src hooks.proto error.proto
