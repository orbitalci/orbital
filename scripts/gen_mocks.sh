#!/usr/bin/env bash
mockgen -source ../models/vcshandler.go -destination ../models/mock_models/vcshandler.mock.go
mockgen -source ../common/credentials/remoteconfig.go -destination ../common/credentials/remoteconfig.mock.go -package credentials
mockgen -source ../storage/storage.go -destination ../storage/storage.mock.go -package storage