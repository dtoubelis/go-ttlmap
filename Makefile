PKG_LIST := $(shell go list ./... | grep -v /vendor/)

lint:
	golangci-lint run

vet:
	go vet $(PKG_LIST)

test:
	go test -v -benchmem -bench=. -short $(PKG_LIST)

race:
	go test -v -race -short $(PKG_LIST)

coverage:
	go test -coverprofile=coverage.txt -covermode=atomic

.PHONY: lint vet sec test race coverage