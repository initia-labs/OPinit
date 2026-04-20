#!/usr/bin/make -f

DOCKER := $(shell which docker)

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.18.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen proto-pulsar-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-pulsar-gen:
	@echo "Generating Dep-Inj Protobuf files"
	@$(protoImage) sh ./scripts/protocgen-pulsar.sh

proto-format:
	@$(protoImage) find ./proto -name "*.proto" -exec buf format {} -w \;

proto-lint:
	@$(protoImage) buf lint --error-format=json ./proto

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

.PHONY: proto-all proto-gen proto-pulsar-gen proto-format proto-lint proto-check-breaking


###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run --timeout=15m --tests=false

lint-fix:
	golangci-lint run --fix --timeout=15m --tests=false

.PHONY: lint lint-fix

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./tests/mocks/*" -not -path "./api/*" -not -name '*.pb.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./tests/mocks/*" -not -path "./api/*" -not -name '*.pb.go' | xargs goimports -w -local github.com/initia-labs/OPinit
.PHONY: format


###############################################################################
###                           Tests 
###############################################################################

test: test-unit

test-all: test-unit test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

test-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

benchmark:
	@go test -timeout 20m -mod=readonly -bench=. ./... 

.PHONY: test test-all test-cover test-unit test-race benchmark
