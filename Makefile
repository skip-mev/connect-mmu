#!/usr/bin/make -f

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_DIR ?= $(CURDIR)/build
build:
	@go build -o $(BUILD_DIR)/ ./...
.PHONY: build

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --out-format=tab
.PHONY: lint

format:
	@find . -name '*.go' -type f -not -path "*/mocks/*" | xargs go run mvdan.cc/gofumpt -w .
	@find . -name '*.go' -type f | xargs go run golang.org/x/tools/cmd/goimports -w -local github.com/skip-mev/connect-mmu
.PHONY: format

mocks:
	@echo "--> generating mocks"
	@go install github.com/vektra/mockery/v2
	@go generate ./...
	@make format
.PHONY: mocks

GOTOOLS=$(shell go list -e -f '{{ join .Imports " "}}' ./tools/tools.go)
tools:
	go install ${GOTOOLS}
.PHONY: tools
###############################################################################
###                                 Install                                 ###
###############################################################################

install-sentry:
	@go install ./validator/cmd/sentry
.PHONY: install-sentry

install-mmu:
	@go install ./cmd/mmu
.PHONY: install-mmu

install: install-mmu install-sentry
.PHONY: install

###############################################################################
###                                 Testing                                 ###
###############################################################################

test:
	@go test ./... -race -v
.PHONY: test

test-e2e:
	@./scripts/setup_dydx_localnet.sh
	cd test && go test -v e2e_test.go
	-cd v4-chain/protocol && make localnet-stop
.PHONY: test-e2e

start-localnet-dydx:
	@./scripts/setup_dydx_localnet.sh
.PHONY: start-localnet-dydx

stop-localnet-dydx:
	@cd v4-chain/protocol && make localnet-stop
.PHONY:  stop-localnet-dydx
