# Rigor Mining

Your dumping ground for research. Rigor Mining is a self-hosted library for
papers and books (PDF/EPUB), with tag-based projects, a built-in reader,
one-click sync to a Kobo, and a browser extension for capturing papers directly
into the application.

## Layout

- `backend/` - Go API server (`gorilla/mux`), SQLite (or Postgres) storage,
  Google OAuth login (or a no-OAuth dev login, see below).
- `frontend/` - React + TypeScript + Tailwind app: library view, item
  detail/metadata editing, PDF/EPUB reader, and the Kobo sync screen.
- `extension/` - a small Chrome/Firefox extension that captures the current
  tab (and its PDF, if it finds one) into your library via an API token.

## Getting Started

The backend and frontend run as two separate processes in local dev - open
two shells.

### Backend

```
cd backend
make install
make start
```

`make install` copies `.env.template` to `.env` (fill in real values there)
and creates `datastore/` for the SQLite file and uploaded library files.
`make start` runs the API on `http://localhost:3000`.

For local dev you don't need real Google OAuth credentials - set
`RM_AUTH_DEV_LOGIN=true` in `.env` (this is the default in the template) and
visit `http://localhost:3000/dev-login` in your browser once to sign in as a
fixed local user. Never set this to `true` in a deployed environment.

### Frontend

```
cd frontend
make install
make start
```

The dev server runs at `http://localhost:8080`. Because it runs on a
different port than the API during local dev, the frontend needs to know
where to send requests - `make start` writes `RIGORMINING_API_BASE=http://localhost:3000`
into `frontend/.env` (gitignored) for you via the `make_dev_env` target.

In production the frontend is built into `frontend/dist` and served
directly by the Go process (`RM_WEB_STATIC_DIR` in the backend `.env`), so
`RIGORMINING_API_BASE` is left empty and requests just go to the same origin.

### Extension

Load `extension/` as an unpacked/temporary extension in Chrome or Firefox,
then open its options page and set the server URL (your backend, e.g.
`http://localhost:3000`) and an API token. Tokens are created on the
Settings page in the app (`/settings`) - they're how the extension
authenticates without doing a browser OAuth round-trip.

## Deploying

There's a repo-root `Dockerfile` and `Makefile` that build both `frontend/`
and `backend/` into one image (`make docker_build`, then `make docker_push`
to Docker Hub - copy `.env.template` to `.env` first and set `REPOSITORY`/
`PROJECT`). Past that point deployment is manual in the Cloud Run console,
point the service at the pushed image tag, override
the `RM_*` env vars under Variables & Secrets with real values (never
`RM_AUTH_DEV_LOGIN=true`), and add a Cloud Storage volume mounted at
`/root/datastore` under Volumes/Volume Mounts. That path is where
`RM_DB_CONNECTION` (the sqlite file) and `RM_LIBRARY_DIR` (uploaded
PDFs/EPUBs) both resolve by default, so mounting a bucket there is what
makes them survive redeploys instead of living on one instance's throwaway
local disk. See the comment above the `docker_run` target in the root
`Makefile` for the exact steps, and the note on `BUCKET` in the root
`.env.template` for why that bucket must not be public.

## What's there today

- **Library**: upload PDFs/EPUBs (or bookmark an item with no file yet and
  attach one later), a left-hand tree of your tags/"projects" to browse and
  filter by, drag-and-drop items onto a tag to file them, bulk-select +
  "add to project", search by title/author.
- **Reader**: in-app PDF and EPUB viewer, page jump/total, download.
- **Sync to Kobo**: one-click copy of selected items onto a mounted Kobo via
  the File System Access API (Chrome/Edge only - Firefox doesn't support it
  yet, so on Firefox use the per-item Download link and copy over USB by
  hand), with the same tag tree to narrow down what you're looking at.
- **Export**: any item can be exported to a Markdown file with frontmatter
  (title/authors/year/doi/tags/notes) for use elsewhere (e.g. Obsidian).

## Not there yet

Kobo highlight/annotation sync (pulling highlights back off the device) is
still just in the desktop application: https://github.com/robrohan/stitch.
Future versions should add that code here.

## License

GPLv3, see [LICENSE](LICENSE).
