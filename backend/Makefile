PHONY: install start build test test-postgres extension skills

include .env
export

# Comes from the root Makefile (see VERSION there).
VERSION ?= $(shell $(MAKE) -s --no-print-directory -C .. version)

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
	cp ../extension/dist/* static/extension/

# Zips downloadable agent skills (see ../.agents/skills/) and stages them
# where RM_SKILLS_DIR (default ./static/skills) expects to find them - the
# "Get Extension" page serves straight out of here in local dev.
skills:
	mkdir -p static/skills
	rm -rf ../.agents/skills/kobo-sync/scripts/__pycache__
	cd ../.agents/skills && zip -r -q kobo-sync.zip kobo-sync
	mv ../.agents/skills/kobo-sync.zip static/skills/kobo-sync.zip

start:
	go run -ldflags "-X main.build=$(VERSION)" cmd/server/main.go

build:
	go build -o rigormining-server -ldflags "-X main.build=$(VERSION)" cmd/server/main.go

test:
	go test ./...

# Runs the repository tests against a throwaway Postgres in Docker, since
# Postgres is the eventual production database and enforces things SQLite
# doesn't (foreign keys, case-sensitive LIKE). The container is removed
# afterwards, pass or fail.
test-postgres:
	@docker run -d --rm --name rm-test-postgres -e POSTGRES_PASSWORD=test -e POSTGRES_DB=rmtest \
		-p 127.0.0.1:55432:5432 postgres:16-alpine >/dev/null
	@trap 'docker stop rm-test-postgres >/dev/null' EXIT; \
	for i in $$(seq 1 60); do docker exec rm-test-postgres pg_isready -U postgres -d rmtest >/dev/null 2>&1 && break; sleep 0.5; done; \
	RM_TEST_POSTGRES="host=127.0.0.1 port=55432 user=postgres password=test dbname=rmtest sslmode=disable" \
		go test -count=1 ./internals/repository/...

# Building/pushing/deploying the container image needs both backend/ and
# frontend/ in the build context - see the Makefile and Dockerfile at the
# repo root (`cd .. && make docker_build`), not here.
