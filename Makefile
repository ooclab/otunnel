# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
# LDFLAGS=-ldflags "-s"
LDFLAGS=-ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`"
STATIC_LDFLAGS=-a -installsuffix cgo -ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`"

PROGRAM_NAME=otunnel
DIST_DIR=dist
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 linux/386 linux/arm linux/mips linux/mipsle windows/amd64 windows/arm64

.PHONY: all static build-all install clean clean-dist

all:
	$(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
static:
	CGO_ENABLED=0 $(GOBUILD) -v $(STATIC_LDFLAGS) -o $(PROGRAM_NAME)

build-all: clean-dist
	@mkdir -p $(DIST_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		EXT=""; \
		EXTRA_ENV=""; \
		if [ "$$GOOS" = "windows" ]; then EXT=".exe"; fi; \
		if [ "$$GOARCH" = "arm" ]; then EXTRA_ENV="GOARM=7"; fi; \
		if [ "$$GOARCH" = "mips" ] || [ "$$GOARCH" = "mipsle" ]; then EXTRA_ENV="GOMIPS=softfloat"; fi; \
		OUTPUT="$(DIST_DIR)/$(PROGRAM_NAME)_$${GOOS}_$${GOARCH}$$EXT"; \
		echo "building $$OUTPUT"; \
		env CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH $$EXTRA_ENV $(GOBUILD) -v $(LDFLAGS) -o "$$OUTPUT" .; \
	done

install:
	$(GOINSTALL) -v

clean:
	@rm -f $(PROGRAM_NAME)*

clean-dist:
	@rm -rf $(DIST_DIR)

