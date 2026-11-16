# Makefile of the Cast package.
#
#   make            same as make build
#   make build      checks (go vet) and builds the package
#   make generate   regenerates cast_switch.go and cast_switch_test.go from cast_gen.go
#   make test       builds the test binary Cast.test and runs it
#   make clean      removes the test binary and the cached test results (go clean -testcache)
#
# Recipes are written on the same line as the target (after ";"), so the Makefile needs no tabs.

GO       ?= go
TEST_BIN := Cast.test

.PHONY: all build generate test clean

all: build

build: ; $(GO) vet ./... && $(GO) build ./...

generate: ; $(GO) generate ./...

test: build ; $(GO) test -c -o $(TEST_BIN) . && ./$(TEST_BIN) -test.count=1 $(TEST_ARGS)

clean: ; rm -f $(TEST_BIN) && $(GO) clean && $(GO) clean -testcache
