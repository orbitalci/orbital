.DEFAULT_GOAL := all
SHELL = sh

.PHONY: help

all: init-googleapi-protos
	cargo build

# Note: VSCode may automatically recreate your cargo target dir, for `rls`.
clean:
	cargo clean
	rm -rf ./models/protos/vendor

init-googleapi-protos:
	if [ ! -d "models/protos/vendor/grpc-gateway" ]; then \
		mkdir -p models/protos/vendor/; \
		cd models/protos/vendor && \
			git clone https://github.com/grpc-ecosystem/grpc-gateway.git; \
	fi

help:
	echo 'Targets:'
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'