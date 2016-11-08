NAME := buckup-to-s3
VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -w -s -X 'main.version=$(VERSION)' -X 'main.revision=$(REVISION)'

## Setup
setup:
	go get github.com/Masterminds/glide

## Run tests
test: deps
	go test $$(glide novendor)

## Resolve dependencies
deps: setup
	glide install

## Update dependencies
update: setup
	glide update

## Build
build: deps
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/backup-to-s3-osx
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/backup-to-s3-linux-amd64

.PHONY: setup test deps update test build
