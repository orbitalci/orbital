.DEFAULT_GOAL := help
SHELL = bash

GOTAGS ?=
GOFILES ?= $(shell go list ./... | grep -v /vendor/)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPATH=$(shell go env GOPATH)

# Get the git commit
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_DESCRIBE=$(shell git describe --tags --always)
GIT_IMPORT=bitbucket.org/level11consulting/ocelot/version
GOLDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) -X $(GIT_IMPORT).GitDescribe=$(GIT_DESCRIBE)

export GOLDFLAGS
SSH_PRIVATE_KEY ?= $(HOME)/.ssh/id_rsa
export SSH_PRIVATE_KEY
GIT_HASH := $(shell git rev-parse --short HEAD)

windows-client: ## install zipped windows ocelot client to pkg/windows_amd64
	mkdir -p pkg/windows_amd64/
	@echo "building windows client"
	env GOOS=windows GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/windows_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/windows_amd64; zip -r ocelot.zip ./ocelot; rm ocelot; cd -

mac-client: ## install zipped mac ocelot client to pkg/darwin_amd64
	mkdir -p pkg/darwin_amd64/
	@echo "building mac client"
	env GOOS=darwin GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/darwin_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/darwin_amd64; zip -r ocelot.zip ./ocelot; rm ocelot; cd -

linux-client: ## install zipped linux ocelot client to pkg/linux_amd64
	mkdir -p pkg/linux_amd64/
	@echo "building mac client"
	env GOOS=linux GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/linux_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/linux_amd64; zip -r ocelot.zip ./ocelot; rm ocelot; cd -

all-clients: windows-client mac-client linux-client ## install all clients

upload-clients: all-clients
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/ocelot.zip s3://ocelotty/mac-ocelot.zip
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/windows_amd64/ocelot.zip s3://ocelotty/windows-ocelot.zip
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/ocelot.zip s3://ocelotty/linux-ocelot.zip

upload-templates: ## tar up werker templates and upload to s3
	cd werker/builder/template; tar -cvf werker_files.tar *
	aws s3 cp --acl public-read-write --content-disposition attachment werker/builder/template/werker_files.tar s3://ocelotty/werker_files.tar
	rm werker/builder/template/werker_files.tar

linux-werker: ## install linux werker zip and upload to s3
	cd cmd/werker/; env GOOS=linux GOARCH=amd64 go build -o werker main.go; zip -r ../../linux-werker.zip werker; rm werker; cd -
	@aws s3 cp --acl public-read-write --content-disposition attachment linux-werker.zip s3://ocelotty/linux-werker.zip
	rm linux-werker.zip

sshexists:
ifeq ("$(wildcard $(SSH_PRIVATE_KEY))","")
	$(error SSH_PRIVATE_KEY must exist or ~/.ssh/id_rsa must exist!)
endif

docker-base: sshexists ## build ocelot-builder base image
	@docker build \
	   --build-arg SSH_PRIVATE_KEY="$$(cat $${SSH_PRIVATE_KEY})" \
	   -f Dockerfile.build \
	   -t ocelot-build \
	   .

docker-build: ## build all images
	@docker-compose build

release: proto upload-clients upload-templates linux-werker docker-base docker-build ## build protos, install & upload clients, upload werker templates, install & upload linux werker, build docker base, build all images

proto: ## build all protos
	@scripts/build-protos.sh

pushtags: ## tag built docker images with the short hash and push all to nexus
	@scripts/tag_and_push.sh $(GIT_HASH)

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'