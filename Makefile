project?=github.com/mprimi/go-bench-away
projectname?=go-bench-away
version?=dev
sha?=$(shell git rev-parse --short HEAD)
date?=$(shell date "+%Y-%m-%d_%H:%M:%S")

default: build

.PHONY: build install run test clean cover vet fmt lint mod update-deps check

build:
	@go build -ldflags "-X $(project)/pkg/core.Version=$(version) -X $(project)/pkg/core.SHA=$(sha) -X $(project)/pkg/core.BuildDate=$(date)" -o $(projectname)

install:
	@go install -ldflags "-X $(project)/pkg/core.Version=$(version) -X $(project)/pkg/core.SHA=$(sha) -X $(project)/pkg/core.BuildDate=$(date)"

run:
	@go run main.go -ldflags "-X $(project)/pkg/core.Version=$(version) -X $(project)/pkg/core.SHA=$(sha) -X $(project)/pkg/core.BuildDate=$(date)" main.go

test:
	@go test -v -failfast -count=1 ./...

clean:
	@rm -rf coverage.out dist/ $(projectname)

cover:
	@go test -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
	@go tool cover -func=coverage.out
	@go tool cover -html coverage.out -o coverage.html

vet:
	@go vet ./...

fmt:
	@gofmt -w -s  .

lint: # Depends on https://github.com/golangci/golangci-lint
	@golangci-lint run -c .golangci-lint.yml --fix

mod:
	@go mod tidy

update-deps:
	@go get -u -t ./...
	@go mod tidy

check: mod fmt test lint cover vet
