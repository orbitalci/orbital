.DEFAULT_GOAL := default
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
GIT_IMPORT=github.com/hashicorp/consul/version
GOLDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) -X $(GIT_IMPORT).GitDescribe=$(GIT_DESCRIBE)

export GOLDFLAGS
SSH_PRIVATE_KEY ?= $(HOME)/.ssh/id_rsa
export SSH_PRIVATE_KEY

windows-client:
	mkdir -p pkg/windows_amd64/
	@echo "building windows client"
	env GOOS=windows GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/windows_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/windows_amd64; zip -r ocelot.zip ./ocelot; rm ocelot; cd -

mac-client:
	mkdir -p pkg/darwin_amd64/
	@echo "building mac client"
	env GOOS=darwin GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/darwin_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/darwin_amd64; zip -r ocelot.zip ./ocelot; rm ocelot; cd -

linux-client:
	mkdir -p pkg/linux_amd64/
	@echo "building mac client"
	env GOOS=linux GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/linux_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/linux_amd64; zip -r ocelot.zip ./ocelot; rm ocelot; cd -

all-clients: windows-client mac-client linux-client

upload-clients:
	all-clients
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/ocelot.zip s3://ocelotty/mac-ocelot.zip
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/windows_amd64/ocelot.zip s3://ocelotty/windows-ocelot.zip
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/ocelot.zip s3://ocelotty/linux-ocelot.zip

upload-templates:
	cd werker/builder/template; tar -cvf werker_files.tar *
	aws s3 cp --acl public-read-write --content-disposition attachment werker/builder/template/werker_files.tar s3://ocelotty/werker_files.tar
	rm werker/builder/template/werker_files.tar

linux-werker:
	echo $(SSH_PRIVATE_KEY)
	cd cmd/werker/; env GOOS=linux GOARCH=amd64 go build -o werker main.go; zip -r ../../linux-werker.zip werker; rm werker; cd -
	@aws s3 cp --acl public-read-write --content-disposition attachment linux-werker.zip s3://ocelotty/linux-werker.zip
	rm linux-werker.zip

sshexists:
ifeq ("$(wildcard $(SSH_PRIVATE_KEY))","")
	$(error SSH_PRIVATE_KEY must exist or ~/.ssh/id_rsa must exist!)
endif

docker-base: sshexists
	@docker build \
	   --build-arg SSH_PRIVATE_KEY="$$(cat $${SSH_PRIVATE_KEY})" \
	   -f Dockerfile.build \
	   -t ocelot-build \
	   .

docker-build:
	@docker-compose build

release: protos upload-clients upload-templates docker-base docker-build

protos:
	scripts/build-protos.sh

default:
	@echo "heeey specify a target please"