.PHONY: list install test docker_build docker_push docker_run clean

HASH=$(shell git log --pretty=format:'%h' -n 1)

include .env
export

DOCKER_CONTAINER=$(REPOSITORY)/$(PROJECT)

# List all targets in this file
list:
	@echo $(HASH)
	@echo ""
	@grep -B 1 '^[^#[:space:]].*:' Makefile

install:
	cd backend && $(MAKE) install
	cd frontend && $(MAKE) install

test:
	cd backend && $(MAKE) test

clean:
	rm -rf frontend/dist backend/rigormining-server

# Build context is the repo root - see Dockerfile.
docker_build:
	docker buildx build --platform linux/amd64 \
		--build-arg VERSION=$(HASH) \
		-t $(DOCKER_CONTAINER):$(HASH) .

docker_push:
	docker push $(DOCKER_CONTAINER):$(HASH)

# Quick local smoke test of the built image - datastore/ isn't mounted
# here, so it just uses the container's local disk (see Dockerfile).
# backend/.env.production is your own gitignored copy of
# backend/.env.template with real (non-dev-login) values.
docker_run:
	docker run --env-file=backend/.env.production -p 8080:3000 $(DOCKER_CONTAINER):$(HASH)

# Deploying past this point is manual:
#   1. `make docker_build docker_push` to get $(DOCKER_CONTAINER):$(HASH)
#      onto Docker Hub.
#   2. Cloud Run console -> the service -> Edit & Deploy New Revision ->
#      Container image URL -> point it at that tag.
#   3. Same screen, Variables & Secrets tab -> override the RM_* vars (see
#      backend/.env.template) with real values - client id/secret, the real
#      redirect URL, allowed origin, etc. Do NOT set RM_AUTH_DEV_LOGIN=true.
#      RM_WEB_STATIC_DIR is already baked into the image, no need to set it.
#   4. Same screen, Volumes tab -> add a Cloud Storage volume for BUCKET
#      (see .env.template - must not be public), then Volume Mounts -> mount
#      it at /root/datastore. That's what RM_DB_CONNECTION and
#      RM_LIBRARY_DIR both resolve under, so this is what makes the sqlite
#      file and every user's uploaded PDFs/EPUBs survive redeploys instead
#      of living on one instance's throwaway local disk.
