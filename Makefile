.PHONY: all clean coverage build test release tools

default: help

help:
	@echo "These 'make' targets are available."
	@echo
	@echo "  test               Run the unit tests"

test:
	go test ./... -v
