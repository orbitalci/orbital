package github

//go:generate  protoc --go_out=plugins=grpc:pb/ -I . -I$GOPATH/src commits.proto hooks.proto general.proto repository.proto error.proto
