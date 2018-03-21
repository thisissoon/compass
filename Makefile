# Compilation Flags
GOOS            ?= $(shell go env GOOS)
GOARCH          ?= $(shell go env GOARCH)
# Build Vars
BUILD_TIME      ?= $(shell date +%s)
BUILD_VERSION   ?= $(shell head -n 1 VERSION | tr -d "\n")
BUILD_COMMIT    ?= $(shell git rev-parse HEAD)
# LDFlags
LDFLAGS ?= -d
LDFLAGS += -X compass/pkg/version.timestamp=$(BUILD_TIME)
LDFLAGS += -X compass/pkg/version.version=$(BUILD_VERSION)
LDFLAGS += -X compass/pkg/version.commit=$(BUILD_COMMIT)
# Go Build Flags
GOBUILD_FLAGS ?= -tags netgo -installsuffix netgo
GOBUILD_FLAGS += -installsuffix netgo
# Bin Dir
BIN_DIR ?= ./_bin
# Compress Binry
COMPRESS_BINARY ?= 0
# Verbose build output
GOBUILD_VERBOSE ?= 0

# Generate protobuf code
.PHONY: protoc
protoc:
	$(MAKE) -C _proto/ all

# Generate database migrations
.PHONY: migrations
migrations:
	$(MAKE) -C _migrations/ all

# Run dep ensure and prun unneeded dependencies
.PHONY: ensure
ensure:
ifeq ("$(wildcard $(shell which dep))","")
	go get github.com/golang/dep/cmd/dep
endif
	dep ensure -v

# Run test suite
test:
ifeq ("$(wildcard $(shell which gocov))","")
	go get github.com/axw/gocov/gocov
endif
	gocov test -v ./... | gocov report

build-flags:
ifeq ($(GOBUILD_VERBOSE),1)
	$(eval GOBUILD_FLAGS += -v)
endif

ldflags:
ifeq ($(COMPRESS_BINARY),1)
	$(eval LDFLAGS += -a -w -s)
endif

compass: build-flags ldflags |
	$(eval BIN_NAME ?= compass.$(BUILD_VERSION).$(GOOS)-$(GOARCH))
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	go build $(GOBUILD_FLAGS) \
		-ldflags "$(LDFLAGS)" \
		-o "$(BIN_DIR)/$(BIN_NAME)" \
		./cmd/compass

compass-image:
	docker build \
		--force-rm \
		--build-arg APP=compass \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
		-t soon/compass:$(BUILD_VERSION) .

needle: build-flags ldflags |
	$(eval BIN_NAME ?= needle.$(BUILD_VERSION).$(GOOS)-$(GOARCH))
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	go build $(GOBUILD_FLAGS) \
		-ldflags "$(LDFLAGS)" \
		-o "$(BIN_DIR)/$(BIN_NAME)" \
		./cmd/needle

needle-image:
	docker build \
		--force-rm \
		--build-arg APP=needle \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
		-t soon/needle:$(BUILD_VERSION) .
