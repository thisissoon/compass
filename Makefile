# Go Binary
GO  ?= go
APP = compass
CGO = 1

# Docker
DOCKER_REGISTRY ?= gcr.io

# Binary Directory
BINDIR ?= $(CURDIR)/_bin

# Go Build Flags
GOFLAGS :=

# Versioning
GIT_COMMIT     ?= $(shell git rev-parse HEAD)
GIT_SHA        ?= $(shell git rev-parse --short HEAD)
GIT_TAG        ?= $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY      ?= $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
ifdef VERSION
	DOCKER_VERSION = $(VERSION)
	BINARY_VERSION = $(VERSION)
endif
DOCKER_VERSION ?= git-${GIT_SHA}
BINARY_VERSION ?= ${GIT_TAG}

# LDFlags
LDFLAGS := -w -s
LDFLAGS += -X compass/pkg/version.Timestamp=$(shell date +%s)
LDFLAGS += -X compass/pkg/version.GitCommit=${GIT_COMMIT}
LDFLAGS += -X compass/pkg/version.GitTreeState=${GIT_DIRTY}
ifneq ($(BINARY_VERSION),)
	LDFLAGS += -X compass/pkg/version.Version=${BINARY_VERSION}
endif

.PHONY: protoc
protoc:
	$(MAKE) -C _proto/ all

.PHONY: migrations
migrations:
	$(MAKE) -C _migrations/ all

.PHONY: ensure
ensure:
ifeq ("$(wildcard $(shell which dep))","")
	go get github.com/golang/dep/cmd/dep
endif
	dep ensure -v

.PHONY: test
test:
ifeq ("$(wildcard $(shell which gocov))","")
	go get github.com/axw/gocov/gocov
endif
	gocov test -v ./... | gocov report

.PHONY: info
info:
	@echo "Version:           ${VERSION}"
	@echo "Git Tag:           ${GIT_TAG}"
	@echo "Git Commit:        ${GIT_COMMIT}"
	@echo "Git Tree State:    ${GIT_DIRTY}"
	@echo "Docker Version:    ${DOCKER_VERSION}"
	@echo "Registry:          ${DOCKER_REGISTRY}"

.PHONY: clean
clean:
	@rm -rf $(BINDIR)

# usage APP=needle make build
.PHONY: build
build:
	CGO_ENABLED=$(CGO) GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -ldflags '$(LDFLAGS)' compass/cmd/$(APP)

# usage APP=needle make static
.PHONY: static
static: GOFLAGS += -a
static: GOFLAGS += -tags netgo -installsuffix netgo
static: GOFLAGS += -installsuffix netgo
static: LDFLAGS += -a -extldflags "-static"
static: CGO = 0
static: build

# usage: APP=needle make image
image:
	docker build \
		--force-rm \
		--build-arg APP=$(APP) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg GIT_SHA=$(GIT_SHA) \
		--build-arg GIT_TAG=$(GIT_TAG) \
		--build-arg GIT_DIRTY=$(GIT_DIRTY) \
		-t soon/$(APP):$(DOCKER_VERSION) .

# usage: make compass
.PHONY: compass
compass: APP = compass
compass: build

# usage: make needle
.PHONY: needle
needle: APP = needle
needle: build
