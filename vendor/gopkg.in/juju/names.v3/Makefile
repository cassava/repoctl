#
# Makefile for juju/names
#

PROJECT := github.com/juju/names/v3
PROJECT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
PROJECT_PACKAGES := $(shell go list $(PROJECT)/... | grep -v /acceptancetests/)
TEST_TIMEOUT := 600s

default: build

build: go-build

# Reformat source files.
format:
	gofmt -w -l .

go-build:
	@go build $(PROJECT_PACKAGES)

test: build
	go test $(CHECK_ARGS) -test.timeout=$(TEST_TIMEOUT) $(PROJECT_PACKAGES) -check.v
