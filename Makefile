.PHONY: test test-integration build docker run tidy

BINARY := bin/gateway

test:
	CGO_ENABLED=0 go test ./...

test-integration:
	go test -tags=integration ./test/...

build:
	go build -o $(BINARY) ./cmd/gateway

docker:
	docker build -f deploy/docker/Dockerfile -t hi-im-gateway:latest .

run: build
	./$(BINARY)

tidy:
	go mod tidy
