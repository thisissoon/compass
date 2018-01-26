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
LDFLAGS += -X compass/version.timestamp=$(BUILD_TIME)
LDFLAGS += -X compass/version.version=$(BUILD_VERSION)
LDFLAGS += -X compass/version.commit=$(BUILD_COMMIT)
# Go Build Flags
GOBUILD_FLAGS ?= -tags netgo -installsuffix netgo
GOBUILD_FLAGS += -installsuffix netgo
# Compress Binry
COMPRESS_BINARY ?= 0
# Verbose build output
GOBUILD_VERBOSE ?= 0

.PHONY: migrations

# Run dep ensure and prun unneeded dependencies
ensure:
ifeq ("$(wildcard $(shell which dep))","")
	go get github.com/golang/dep/cmd/dep
endif
	dep ensure -v

# Compile protobuf to go
protoc:
ifeq ("$(wildcard $(shell which protoc-gen-go))","")
	go get github.com/golang/protobuf/protoc-gen-go
endif
	protoc -I .:/usr/local/include --go_out=plugins=grpc:./proto $(shell find . -type f -name '*.proto')

# Generate migration code
migrations:
ifeq ("$(wildcard $(shell which go-bindata))","")
	go get github.com/jteeuwen/go-bindata/go-bindata
endif
	go-bindata -pkg=migrations -prefix=migrations/ -o=store/psql/migrations/migrations.go ./migrations

# Run test suite
test:
ifeq ("$(wildcard $(shell which gocov))","")
	go get github.com/axw/gocov/gocov
endif
	ADMINAUTHENTICATOR_LOG_FORMAT=discard \
		gocov test -v ./... | gocov report

common-build-flags:
ifeq ($(GOBUILD_VERBOSE),1)
	$(eval GOBUILD_FLAGS += -v)
endif

common-ldflags:
ifeq ($(COMPRESS_BINARY),1)
	$(eval LDFLAGS += -a -w -s)
endif

compass-ldflags:
	$(eval LDFLAGS += -X compass/config.filename=compass)
	$(eval LDFLAGS += -X compass/config.envprefix=compass)

compass: common-build-flags common-ldflags compass-ldflags |
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

needle-ldflags:
	$(eval LDFLAGS += -X compass/config.filename=needle)
	$(eval LDFLAGS += -X compass/config.envprefix=needle)

needle: common-build-flags common-ldflags needle-ldflags |
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
