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
GIT_IMPORT=github.com/level11consulting/orbitalci/version
GOLDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) -X $(GIT_IMPORT).GitDescribe=$(GIT_DESCRIBE)
GOLDFLAGS_REL=$(GOLDFLAGS) -X $(GIT_IMPORT).VersionPrerelease=
export GOLDFLAGS
GIT_HASH := $(shell git rev-parse --short HEAD)

versionexists:
ifndef VERSION
	$(error VERSION must be applied by maket target VERSION=x or another method if building/uploading clients clients)
endif


# This needs a build w/o an installation


local: ## install locally but with the tags/flags injected in
	go install -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' ./...

local-release:
	go install -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' ./...

local-service:
	go install -ldflags '$(GOLDFLAGS_REL)' -tags '$(GOTAGS)' -a ./cmd/$(SERVICE_NAME)

legacy-windows-client: versionexists ## install zipped windows ocelot client to pkg/windows_amd64
	mkdir -p pkg/windows_amd64/
	@echo "building windows client"
	env GOOS=windows GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/windows_amd64/ocelot.exe  cmd/ocelot/main.go
	cd pkg/windows_amd64; zip -r ocelot_$(VERSION).zip ./ocelot.exe; rm ocelot.exe; cd -

legacy-mac-client: versionexists ## install zipped mac ocelot client to pkg/darwin_amd64
	mkdir -p pkg/darwin_amd64/
	@echo "building mac client"
	env GOOS=darwin GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/darwin_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/darwin_amd64; zip -r ocelot_$(VERSION).zip ./ocelot; rm ocelot; cd -

legacy-linux-client: versionexists ## install zipped linux ocelot client to pkg/linux_amd64
	mkdir -p pkg/linux_amd64/
	@echo "building mac client"
	env GOOS=linux GOARCH=amd64 go build  -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' -o pkg/linux_amd64/ocelot  cmd/ocelot/main.go
	cd pkg/linux_amd64; zip -r ocelot_$(VERSION).zip ./ocelot; rm ocelot; cd -

all-clients: legacy-windows-client legacy-mac-client legacy-linux-client ## install all clients

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

docker-build:
	@docker build \
	   -f Dockerfile \
	   -t orbitalci \
	   .

start-docker-infra:
	cd deploy/infra; \
	docker-compose up -d

stop-docker-infra:
	cd deploy/infra; \
	docker-compose down

start-docker-orbital: start-docker-infra
	docker-compose up -d

stop-docker-orbital:
	docker-compose down 

#release: proto upload-clients upload-templates linux-werker docker-base docker-build ## build protos, install & upload clients, upload werker templates, install & upload linux werker, build docker base, build all images

proto: ## build all protos
	go generate ./models/...

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
