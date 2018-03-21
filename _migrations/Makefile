# Generate migration code
.PHONY: all
all:
ifeq ("$(wildcard $(shell which go-bindata))","")
	go get github.com/containous/go-bindata/go-bindata
endif
	go-bindata -pkg=psql -prefix=migrations/ -o=../store/psql/migrations.go .

