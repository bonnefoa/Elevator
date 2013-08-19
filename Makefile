SHELL := /usr/bin/env bash
ELEVATOR_PACKAGE := github.com/oleiade/Elevator

BUILD_DIR := $(CURDIR)/.gopath

DEPS_DIR := $(CURDIR)/deps
LEVELDB_DIR := $(DEPS_DIR)/leveldb

GOPATH ?= $(BUILD_DIR)
export GOPATH

GO_OPTIONS ?=
ifeq ($(VERBOSE), 1)
GO_OPTIONS += -v
endif

GIT_COMMIT = $(shell git rev-parse --short HEAD)
GIT_STATUS = $(shell test -n "`git status --porcelain`" && echo "+CHANGES")

NO_MEMORY_LIMIT ?= 0
export NO_MEMORY_LIMIT

BUILD_OPTIONS = -ldflags "-X main.GIT_COMMIT $(GIT_COMMIT)$(GIT_STATUS) -X main.NO_MEMORY_LIMIT $(NO_MEMORY_LIMIT)"

SRC_DIR := $(GOPATH)/src

ELEVATOR_DIR := $(SRC_DIR)/$(ELEVATOR_PACKAGE)

ELEVATOR_MAIN := $(ELEVATOR_DIR)/elevator
ELEVATOR_BIN_RELATIVE := bin/elevator
ELEVATOR_BIN := $(CURDIR)/$(ELEVATOR_BIN_RELATIVE)

ELEVATOR_CLI_MAIN := $(ELEVATOR_DIR)/cli
ELEVATOR_CLI_BIN_RELATIVE := bin/cli
ELEVATOR_CLI_BIN := $(CURDIR)/$(ELEVATOR_CLI_BIN_RELATIVE)

SRCS = $(wildcard **/*.go)

.PHONY: all clean test

all: $(ELEVATOR_BIN) $(ELEVATOR_CLI_BIN) 

$(ELEVATOR_BIN): $(ELEVATOR_DIR) store/request.pb.go server/response.pb.go
	# Proceed to elevator build
	@(mkdir -p  $(dir $@))
	@(cd $(ELEVATOR_MAIN); go get $(GO_OPTIONS); go build $(GO_OPTIONS) $(BUILD_OPTIONS) -o $@)
	@echo $(ELEVATOR_BIN_RELATIVE) is created.

$(ELEVATOR_CLI_BIN): $(ELEVATOR_DIR)
	@(mkdir -p  $(dir $@))
	@(cd $(ELEVATOR_CLI_MAIN); go get $(GO_OPTIONS); go build $(GO_OPTIONS) $(BUILD_OPTIONS) -o $@)
	@echo $(ELEVATOR_CLI_BIN_RELATIVE) is created.

$(ELEVATOR_DIR): $(SRCS)
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR)/ $@

clean:
ifeq ($(GOPATH), $(BUILD_DIR))
	@rm -rf $(BUILD_DIR)
else ifneq ($(ELEVATOR_DIR), $(realpath $(ELEVATOR_DIR)))
	@rm -f $(ELEVATOR_DIR)
endif

PACKAGES := $(shell find . -iname '*_test.go' | xargs -I{} dirname {} | sort | uniq | sed -rn 's`\.`$(ELEVATOR_PACKAGE)`p')

test: all
	@go test $(GO_OPTIONS) $(PACKAGES)

bench: all
	@go test $(PACKAGES) -bench .

fmt:
	@gofmt -s -l -w .

%.pb.go:
	protoc --go_out=. $*.proto
