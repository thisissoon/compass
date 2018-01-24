# Compilation Flags
GOOS            ?= $(shell go env GOOS)
GOARCH          ?= $(shell go env GOARCH)
# Flags
FLAGS           ?=
# Build Vars
BUILD_TIME      ?= $(shell date +%s)
BUILD_VERSION   ?= $(shell head -n 1 VERSION | tr -d "\n")
BUILD_COMMIT    ?= $(shell git rev-parse HEAD)
# LDFlags
LDFLAGS ?= -d
LDFLAGS += -X compass/config.timestamp=$(BUILD_TIME)
LDFLAGS += -X compass/config.version=$(BUILD_VERSION)
LDFLAGS += -X compass/config.commit=$(BUILD_COMMIT)
# Go Build Flags
GOBUILD_FLAGS ?= -tags netgo -installsuffix netgo
GOBUILD_FLAGS += -installsuffix netgo
# Compress Binry
COMPRESS_BINARY ?= 0
# Verbose build output
GOBUILD_VERBOSE ?= 0

# Run dep ensure and prun unneeded dependencies
ensure:
ifeq ("$(wildcard $(shell which dep))","")
	go get github.com/golang/dep/cmd/dep
endif
	dep ensure -v && dep prune -v

common-build-flags:
ifeq ($(GOBUILD_VERBOSE),1)
	$(eval GOBUILD_FLAGS += -v)
endif

common-ldflags:
ifeq ($(COMPRESS_BINARY),1)
	$(eval LDFLAGS += -a -w -s)
endif

compass: common-build-flags common-ldflags |
	$(eval BIN_NAME ?= compass.$(BUILD_VERSION).$(GOOS)-$(GOARCH))
	CGO_ENABLED=0 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	go build $(GOBUILD_FLAGS) \
		-ldflags "$(LDFLAGS)" \
		-o "$(BIN_NAME)" \
		./cmd/compass

compass-image:
	docker build \
		--force-rm \
		--build-arg APP=compass \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
		-t soon/compass:$(BUILD_VERSION) .

needle: common-build-flags common-ldflags |
	$(eval BIN_NAME ?= needle.$(BUILD_VERSION).$(GOOS)-$(GOARCH))
	CGO_ENABLED=0 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	go build $(GOBUILD_FLAGS) \
		-ldflags "$(LDFLAGS)" \
		-o "$(BIN_NAME)" \
		./cmd/needle

needle-image:
	docker build \
		--force-rm \
		--build-arg APP=needle \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
		-t soon/needle:$(BUILD_VERSION) .

# Run test suite
test:
ifeq ("$(wildcard $(shell which gocov))","")
	go get github.com/axw/gocov/gocov
endif
	ADMINAUTHENTICATOR_LOG_FORMAT=discard \
		gocov test -v ./... | gocov report
