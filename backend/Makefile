PHONY: install start build test extension skills

include .env
export

install: extension skills
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

# Zips downloadable agent skills (see ../.agents/skills/) and stages them
# where RM_SKILLS_DIR (default ./static/skills) expects to find them - the
# "Get Extension" page serves straight out of here in local dev.
skills:
	mkdir -p static/skills
	rm -rf ../.agents/skills/kobo-sync/scripts/__pycache__
	cd ../.agents/skills && zip -r -q kobo-sync.zip kobo-sync
	mv ../.agents/skills/kobo-sync.zip static/skills/kobo-sync.zip

start:
	go run cmd/server/main.go

build:
	go build -o rigormining-server cmd/server/main.go

test:
	go test ./...

# Building/pushing/deploying the container image needs both backend/ and
# frontend/ in the build context - see the Makefile and Dockerfile at the
# repo root (`cd .. && make docker_build`), not here.
