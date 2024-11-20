#!/usr/bin/make -f

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_DIR ?= $(CURDIR)/build
build:
	@go build -o $(BUILD_DIR)/ ./...

.PHONY: build


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