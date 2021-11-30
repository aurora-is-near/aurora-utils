.PHONY: build test test-install

build:
	env GO111MODULE=on go build -v ./...

test:
	gocheck -c

test-install:
	go get github.com/frankbraun/gocheck
	go get golang.org/x/tools/cmd/goimports
	go get golang.org/x/lint/golint
