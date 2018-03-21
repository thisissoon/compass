import_path = compass/pkg/proto

dst 	= ../pkg/proto
target  = go
plugins = grpc

services_pbs = $(sort $(wildcard services/*.proto))

.PHONY: all
all: services

.PHONY: services
services:
	protoc -I .:/usr/local/include --go_out=plugins=$(plugins):$(dst) $(services_pbs)

.PHONY: clean
clean:
	@rm -rf $(dst)/services 2>/dev/null
