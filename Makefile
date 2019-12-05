.DEFAULT_GOAL := all
SHELL = sh

.PHONY: help

all:
	cargo build

release:
	cargo build --release

docs:
	cargo doc

lint:
	cargo clippy 

# Note: VSCode may automatically recreate your cargo target dir, for `rls`.
clean:
	cargo clean;
	rm -rf ./models/protos/vendor;

help:
	echo 'Targets:'
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'