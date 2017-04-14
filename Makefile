PHONY: test lint

test:
	go test -cover ./...

fmt:
	go fmt ./...

install:
	go install
