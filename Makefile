.PHONY: help test

default: help

help:
	@echo "These 'make' targets are available."
	@echo
	@echo "  test               Run the unit tests"

tools:
	go get github.com/xeipuuv/gojsonschema
	go get github.com/stretchr/testify

test: tools
	go test ./... -v
