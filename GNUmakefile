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
GIT_IMPORT=github.com/shankj3/ocelot/version
GOLDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) -X $(GIT_IMPORT).GitDescribe=$(GIT_DESCRIBE)
GOLDFLAGS_REL=$(GOLDFLAGS) -X $(GIT_IMPORT).VersionPrerelease=
export GOLDFLAGS
GIT_HASH := $(shell git rev-parse --short HEAD)

versionexists:
ifndef VERSION
	$(error VERSION must be applied by maket target VERSION=x or another method if building/uploading clients clients)
endif


local: ## install locally but with the tags/flags injected in
	go install -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' ./...

local-release:
	go install -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' ./...

local-service:
	go install -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -a ./cmd/$(SERVICE_NAME)

windows-client: versionexists ## install zipped windows ocelot client to pkg/windows_amd64
	mkdir -p pkg/windows_amd64/
	@echo "building windows client"
	env GOOS=windows GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/windows_amd64/ocelot.exe  cmd/ocelot/main.go
	cd pkg/windows_amd64; zip -r ocelot_$(VERSION).zip ./ocelot.exe; rm ocelot.exe; cd -

mac-client: versionexists ## install zipped mac ocelot client to pkg/darwin_amd64
	mkdir -p pkg/darwin_amd64/
	@echo "building mac client"
	env GOOS=darwin GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/darwin_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/darwin_amd64; zip -r ocelot_$(VERSION).zip ./ocelot; rm ocelot; cd -

linux-client: versionexists ## install zipped linux ocelot client to pkg/linux_amd64
	mkdir -p pkg/linux_amd64/
	@echo "building mac client"
	env GOOS=linux GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/linux_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/linux_amd64; zip -r ocelot_$(VERSION).zip ./ocelot; rm ocelot; cd -

all-clients: windows-client mac-client linux-client ## install all clients

all-clients-latest: all-clients  ## upload all clients to s3 without a version. VERSION is still required, because idk but it won't be in s3
	 @aws s3 cp --acl public-read-write --content-disposition attachment pkg/darwin_amd64/ocelot_$(VERSION).zip s3://ocelotty/mac-ocelot.zip
	 @aws s3 cp --acl public-read-write --content-disposition attachment pkg/windows_amd64/ocelot_$(VERSION).zip s3://ocelotty/windows-ocelot.zip
     @aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/ocelot_$(VERSION).zip s3://ocelotty/linux-ocelot.zip

build-upload-clients: versionexists all-clients upload-clients ## install all clients and upload to s3

upload-clients: versionexists ## upload already build clients in pkg/os_amd64/ocelot_<version>.zip to s3
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/darwin_amd64/ocelot_$(VERSION).zip s3://ocelotty/mac-ocelot-$(VERSION).zip
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/windows_amd64/ocelot_$(VERSION).zip s3://ocelotty/windows-ocelot-$(VERSION).zip
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/ocelot_$(VERSION).zip s3://ocelotty/linux-ocelot-$(VERSION).zip

all-binaries-rel: versionexists ## build all binaries in RELEASE MODE and save them to pkg
	# darwin client
	env GOOS=darwin GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -o pkg/darwin_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/darwin_amd64; zip -r darwin-ocelot-$(VERSION).zip ./ocelot; rm ocelot; cd -
	# windows client
	env GOOS=windows GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -o pkg/windows_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/windows_amd64; zip -r windows-ocelot-$(VERSION).zip ./ocelot; rm ocelot; cd -
	# linux client
	env GOOS=linux GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -o pkg/linux_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/linux_amd64; zip -r linux-ocelot-$(VERSION).zip ./ocelot; rm ocelot; cd -
	# werker linux
	cd cmd/werker/; env GOOS=linux GOARCH=amd64 go build -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -o werker .; zip -r ../../pkg/linux_amd64/linux-werker-$(VERSION).zip werker; rm werker; cd -
	# werker darwin
	cd cmd/werker/; env GOOS=darwin GOARCH=amd64 go build -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -o werker .; zip -r ../../pkg/darwin_amd64/darwin-werker-$(VERSION).zip werker; rm werker; cd -
	# hookhandler linux
	cd cmd/hookhandler/; env GOOS=linux GOARCH=amd64 go build -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -o hookhandler .; zip -r ../../pkg/linux_amd64/linux-hookhandler-$(VERSION).zip hookhandler; rm hookhandler; cd -

upload-binaries: versionexists ## upload all built binaries
	# upload clients
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/darwin_amd64/darwin-ocelot-$(VERSION).zip s3://ocelotty/
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/windows_amd64/windows-ocelot-$(VERSION).zip s3://ocelotty/
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/linux-ocelot-$(VERSION).zip s3://ocelotty/
	# upload werkers
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/linux_amd64/linux-werker-$(VERSION).zip s3://ocelotty/
	@aws s3 cp --acl public-read-write --content-disposition attachment pkg/darwin_amd64/darwin-werker-$(VERSION).zip s3://ocelotty/

upload-templates: ## tar up werker templates and upload to s3
	cd build/template && tar -cvf werker_files.tar *
	aws s3 cp --acl public-read-write --content-disposition attachment build/template/werker_files.tar s3://ocelotty/werker_files.tar
	rm build/template/werker_files.tar

dev-templates: ## tar up templates and move tarball to cmd/werker/dev
	cd build/template && tar -cvf werker_files.tar *
	mv build/template/werker_files.tar router/werker/werker_files.tar

linux-werker: versionexists ## install linux werker zip and upload to s3
	cd cmd/werker/; env GOOS=linux GOARCH=amd64 go build -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o werker .; zip -r ../../linux-werker-$(VERSION).zip werker; rm werker; cd -
	@aws s3 cp --acl public-read-write --content-disposition attachment linux-werker-$(VERSION).zip s3://ocelotty/linux-werker-$(VERSION).zip
	rm linux-werker-$(VERSION).zip

linux-hookhandler: versionexists ## install linux hookhandler zip and upload to s3
	cd cmd/hookhandler/; env GOOS=linux GOARCH=amd64 go build -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o hookhandler .; zip -r ../../linux-hookhandler-$(VERSION).zip hookhandler; rm hookhandler; cd -
	@aws s3 cp --acl public-read-write --content-disposition attachment linux-hookhandler-$(VERSION).zip s3://ocelotty/linux-hookhandler-$(VERSION).zip
	rm linux-hookhandler-$(VERSION).zip

darwin-werker: versionexists ## install mac werker zip and upload to s3
	cd cmd/werker/; env GOOS=darwin GOARCH=amd64 go build -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o werker .; zip -r ../../darwin-werker-$(VERSION).zip werker; rm werker; cd -
	@aws s3 cp --acl public-read-write --content-disposition attachment darwin-werker-$(VERSION).zip s3://ocelotty/darwin-werker-$(VERSION).zip
	rm darwin-werker-$(VERSION).zip

docker-base: ## build ocelot-builder base image
	@docker build \
	   -f Dockerfile.build \
	   -t ocelot-build \
	   .

docker-build: ## build all images
	@docker-compose build

release: proto upload-clients upload-templates linux-werker docker-base docker-build ## build protos, install & upload clients, upload werker templates, install & upload linux werker, build docker base, build all images

proto: ## build all protos
	go generate ./models/...

pushtags: versionexists ## tag built docker images with the VERSION and push all to nexus
	@scripts/tag_and_push.sh $(VERSION)

admintagpush: versionexists ## tag and push admin docker image
	 docker tag ocelot-admin docker.metaverse.l11.com/ocelot-admin:$(VERSION)
	 docker push docker.metaverse.l11.com/ocelot-admin:$(VERSION)

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
