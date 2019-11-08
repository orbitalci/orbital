.DEFAULT_GOAL := all
SHELL = sh

.PHONY: help

all: init-protobuf-types
	cargo build

release: init-protobuf-types
	cargo build --release

# Note: VSCode may automatically recreate your cargo target dir, for `rls`.
clean:
	cargo clean;
	rm -rf ./models/protos/vendor;

init-protobuf-types:
	if [ ! -d "models/protos/vendor/protobuf" ]; then \
		mkdir -p models/protos/vendor/; \
		cd models/protos/vendor && \
			git clone https://github.com/protocolbuffers/protobuf.git; \
	fi

help:
	echo 'Targets:'
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'