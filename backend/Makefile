PHONY: install start build test extension

include .env
export

install: extension
	go mod tidy
	cp -n .env.template .env || true
	mkdir -p datastore/library

# Builds the browser extension zips (see extension/Makefile) and stages
# them where RM_EXTENSION_DIR (default ./static/extension) expects to find
# them - the "Get Extension" page serves straight out of here in local dev.
extension:
	cd ../extension && $(MAKE) build
	mkdir -p static/extension
	cp ../extension/dist/*.zip static/extension/

start:
	go run cmd/server/main.go

build:
	go build -o rigormining-server cmd/server/main.go

test:
	go test ./...

# Building/pushing/deploying the container image needs both backend/ and
# frontend/ in the build context - see the Makefile and Dockerfile at the
# repo root (`cd .. && make docker_build`), not here.
