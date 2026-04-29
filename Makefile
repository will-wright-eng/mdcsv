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
	@set -e; bin=$(GOBASE)/dist/$(BINARY_NAME); \
	echo "md→csv (file in, stdout)";        $$bin -f md -t csv testdata/simple.md  | diff testdata/simple.csv -; \
	echo "csv→md (file in, stdout)";        $$bin -f csv -t md testdata/simple.csv | diff testdata/simple.md  -; \
	echo "md→md (reformat messy)";          $$bin -f md -t md  testdata/messy.md   | diff testdata/simple.md  -; \
	echo "md→csv (stdin pipe)";             cat testdata/simple.md  | $$bin -f md -t csv | diff testdata/simple.csv -; \
	echo "extension inference (in + -o)";   $$bin testdata/simple.md -o /tmp/mdcsv-smoke.csv && diff testdata/simple.csv /tmp/mdcsv-smoke.csv; \
	echo "extension inference (-o only, stdin)"; cat testdata/simple.md | $$bin -f md -o /tmp/mdcsv-smoke.csv && diff testdata/simple.csv /tmp/mdcsv-smoke.csv; \
	echo "stdin without -f errors";         if cat testdata/simple.md | $$bin -t csv 2>/dev/null; then echo "  expected non-zero exit" >&2; exit 1; fi; \
	rm -f /tmp/mdcsv-smoke.csv; \
	echo "all smoke checks passed"

#* Install Commands
install: build ## install binary
	sudo mkdir -p $(INSTALL_PATH)
	sudo cp $(GOBASE)/dist/$(BINARY_NAME) $(INSTALL_PATH)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)

uninstall: ## uninstall binary
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
