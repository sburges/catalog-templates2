.PHONY: help test

default: help

help:
	@echo "These 'make' targets are available."
	@echo
	@echo "  test               Run the unit tests"

all: tools test

tools:
	@echo "$(OK_COLOR)==> Getting required tools$(NO_COLOR)"
	go get github.com/tools/godep

test: tools test-format
	@echo "$(OK_COLOR)==> Testing code$(NO_COLOR)"
	godep go test ./... -v

test-format:
	@echo "$(OK_COLOR)==> Checking code with gofmt$(NO_COLOR)"
	./scripts/testFmt.sh tests
