# Makefile of the async package.
#
#   make            same as make build
#   make build      checks (go vet) and builds the package
#   make test       builds the test binary async.test (with -race) and runs it
#   make bench      runs the benchmarks
#   make clean      removes the test binary and the cached test results (go clean -testcache)
#
#   make test TEST_ARGS="-test.v -test.run=WorkLock"   passes arguments to the test binary
#
# Recipes are written on the same line as the target (after ";"), so the Makefile needs no tabs.
#
# GOWORK=off: the module is built as users get it (as in CI), independent of a local go.work.
# CGO_ENABLED=1: required by -race.

GO       ?= go
TEST_BIN := async.test

export GOWORK ?= off

.PHONY: all build test bench clean

all: build

build: ; $(GO) vet ./... && $(GO) build ./...

test: build ; CGO_ENABLED=1 $(GO) test -race -c -o $(TEST_BIN) . && ./$(TEST_BIN) -test.count=1 $(TEST_ARGS)

bench: build ; $(GO) test -run=^$$ -bench=. -benchmem .

clean: ; rm -f $(TEST_BIN) && $(GO) clean && $(GO) clean -testcache
