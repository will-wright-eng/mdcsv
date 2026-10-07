# Build configuration
BINARY_NAME=mdcsv
INSTALL_PATH=/usr/local/bin

# Go settings
GOBASE=$(shell pwd)
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

#* Setup
.PHONY: $(shell sed -n -e '/^$$/ { n ; /^[^ .\#][^ ]*:/ { s/:.*$$// ; p ; } ; }' $(MAKEFILE_LIST))
.DEFAULT_GOAL := help

help: ## list make commands
	@echo ${MAKEFILE_LIST}
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

#* Go Commands
build: ## build binary
	mkdir -p $(GOBASE)/dist
	$(GOBUILD) -o $(GOBASE)/dist/$(BINARY_NAME) ./main.go

clean: ## remove binary
	rm -f $(GOBASE)/dist/$(BINARY_NAME)

test: ## run unit tests
	$(GOTEST) ./...

fmt: ## format code
	gofmt -w .

#* End-to-end smoke tests
smoke: build ## run end-to-end smoke tests against testdata fixtures
	@set -e; bin=$(GOBASE)/dist/$(BINARY_NAME); tmp=$$(mktemp -d); \
	echo "md→csv (file in, stdout)";        $$bin testdata/simple.md  | diff testdata/simple.csv -; \
	echo "csv→md (file in, stdout)";        $$bin testdata/simple.csv | diff testdata/simple.md  -; \
	echo "md→md (reformat via -o .md)";     $$bin testdata/messy.md -o $$tmp/clean.md && diff testdata/simple.md $$tmp/clean.md; \
	echo "md→csv (stdin pipe, sniffed)";    cat testdata/simple.md  | $$bin | diff testdata/simple.csv -; \
	echo "csv→md (stdin pipe, sniffed)";    cat testdata/simple.csv | $$bin | diff testdata/simple.md  -; \
	echo "md→csv (file in, -o .csv)";       $$bin testdata/simple.md -o $$tmp/out.csv && diff testdata/simple.csv $$tmp/out.csv; \
	echo "unknown input extension errors";  if $$bin go.mod 2>/dev/null; then echo "  expected non-zero exit" >&2; exit 1; fi; \
	rm -rf $$tmp; \
	echo "all smoke checks passed"

#* Install Commands
install: build ## install binary
	sudo mkdir -p $(INSTALL_PATH)
	sudo cp $(GOBASE)/dist/$(BINARY_NAME) $(INSTALL_PATH)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)

uninstall: ## uninstall binary
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
