SHELL := /bin/bash
# Go parameters
BINARY_NAME=upload-bridge
BINARY_UNIX=$(BINARY_NAME)_unix
REPO=ghcr.io/dathan/go-upload-bridge-pinata

.PHONY: all
all: lint test build

.PHONY: lint
lint:
				golangci-lint run ./...

.PHONY: build
build:
				go build -o ./bin ./cmd/...

.PHONY: test
test:
				go test -p 6 -covermode=count -coverprofile=test/coverage.out test/*.go

.PHONY: clean
clean:
				go clean
				find . -type d -name '.tmp_*' -prune -exec rm -rvf {} \;

.PHONY: run
run:
				source .env && go run ./cmd/$(BINARY_NAME)/*.go

.PHONY: vendor
vendor:
				go mod vendor

# Cross compilation
.PHONY: build-linux
build-linux:
				CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_UNIX) -v cmd/$(BINARY_NAME)/

.PHONY: docker-login
docker-login:
				echo $(CR_PAT) |docker login ghcr.io -u dathan --password-stdin
# Build docker containers
.PHONY: docker-build
docker-build:
				DOCKER_BUILDKIT=1 docker buildx build --platform linux/amd64 -t $(REPO):latest --push .

.PHONY: docker-tag
docker-tag:
			docker tag `docker image ls --filter 'reference=$(BINARY_NAME)-release' -q` $(REPO):latest

# Push the container
.PHONY: docker-push
docker-push: docker-build docker-tag
				docker push $(REPO):latest


.PHONY: docker-clean
docker-clean:
				docker rmi `docker image ls --filter 'reference=$(BINARY_NAME)-*' -q`

.PHONY: docker-run
docker-run:
				docker run --env-file=.env -p 8181:8181 $(REPO):latest