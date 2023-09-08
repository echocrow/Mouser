HASH := $(shell git rev-parse --short HEAD)
COMMIT_DATE := $(shell git show -s --format=%ci ${HASH})
BUILD_DATE := $(shell date '+%Y-%m-%d %H:%M:%S')
VERSION := ${HASH} (${COMMIT_DATE})

.PHONY: run
run:
	@go run -race ./cmd/mouser

.PHONY: mock
mock:
	# Deleting old mocks...
	@find . -name mocks -type d -print0|xargs -0 rm -r --
	# Generating new mocks...
	@go generate ./...

.PHONY: test
test:
	@go test -race -timeout 1s ./...

.PHONY: build
build:
	@goreleaser --snapshot --skip-publish --clean
