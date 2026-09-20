# Build context is the repo root (`docker build .` from here), since the
# image needs both frontend/ and backend/ - see Makefile's docker_build.

# --- frontend: build the static SPA bundle ---
FROM node:20 AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ ./
ARG VERSION=docker
# Mirrors frontend/Makefile's make_prod_env target, without needing a git
# checkout in the build context - VERSION is passed in as a build arg
# instead (see Makefile's docker_build: --build-arg VERSION=$(HASH)).
RUN printf 'RIGORMINING_BUCKET=https://rigormining.com\nRIGORMINING_VERSION=%s\n' "$VERSION" > .env \
    && npm run build \
    && cp node_modules/pdfjs-dist/build/pdf.worker.js dist/pdf.worker.js

# --- extension: zip the browser extension for the "Get Extension" page ---
FROM alpine:3 AS extension-builder
WORKDIR /app/extension
RUN apk add --no-cache zip make
COPY extension/ ./
RUN make build

# --- skills: zip downloadable agent skills for the "Get Extension" page ---
FROM alpine:3 AS skills-builder
WORKDIR /app/skills
RUN apk add --no-cache zip
COPY .agents/skills/kobo-sync ./kobo-sync
RUN zip -r -q kobo-sync.zip kobo-sync

# --- backend: build the Go binary ---
FROM golang:1.25 AS backend-builder
WORKDIR /go/src/gitlab.com/robrohan/rigormining
COPY backend/ ./
ARG VERSION=docker
# CGO must stay on - go-sqlite3 needs it.
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o rigormining-server -ldflags "-X main.build=${VERSION}" cmd/server/main.go

# --- runtime ---
# Same base image as the builder (not a slim/alpine image) so the CGO
# sqlite3 binding doesn't hit a glibc/musl mismatch at runtime.
FROM golang:1.25
WORKDIR /root/
COPY --from=backend-builder /go/src/gitlab.com/robrohan/rigormining/rigormining-server ./
COPY --from=backend-builder /go/src/gitlab.com/robrohan/rigormining/migrations ./migrations
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
COPY --from=extension-builder /app/extension/dist ./static/extension
COPY --from=skills-builder /app/skills/kobo-sync.zip ./static/skills/kobo-sync.zip

# Falls back to local container disk if nothing is mounted at ./datastore -
# fine for a quick `make docker_run` smoke test, but on Cloud Run this path
# must have the Cloud Storage volume mounted over it manually in the
# console (see the comment above docker_run in the root Makefile), since
# Cloud Run instances are ephemeral/stateless and this is where the sqlite
# db and every uploaded PDF/EPUB live.
RUN mkdir -p datastore/library
ENV RM_WEB_STATIC_DIR=./frontend/dist

EXPOSE 3000
CMD ["./rigormining-server"]
