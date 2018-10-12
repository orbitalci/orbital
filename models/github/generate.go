package github

//go:generate  protoc --go_out=plugins=grpc:pb/ -I . -I$GOPATH/src commits.proto hooks.proto owner.proto repository.proto error.proto
