---
name: rigormining
description: Sync a Rigor Mining library with a locally mounted Kobo e-reader over the app's REST API - copies items queued as request_sync onto the device, deletes items queued as request_remove, and reports back what happened. Also adds papers to the library by arXiv id or URL. Invoke when the user asks to add/save/send a paper to Rigor Mining (their library or research server), to sync items to their Kobo, push papers/books to their e-reader, remove synced items from a plugged-in Kobo, or check what's queued to sync.
---

# Rigor Mining

Syncs whatever is currently queued in a Rigor Mining library (see the
"Sync to Kobo" screen in the app, or `backend/cmd/server/main.go`'s
`/api/v1/items/{id}/sync*` routes) onto a Kobo e-reader that's physically
plugged into this machine right now. It can also add papers to the
library (see "Adding papers" below). All the actual work is done by
`scripts/rigormining.py` (stdlib-only Python, no dependencies) - this file
just explains when and how to call it.

Nothing in this skill is tied to a specific server, account, or computer.
Config (server URL + API token) is asked for once and cached outside the
repo at `~/.rigormining-sync.json` (mode 0600) - never read that file's
contents back to the user, never write the token anywhere else, and never
put a real server URL or token into this skill directory itself.

## 1. Setup (only if not already configured)

Check whether `~/.rigormining-sync.json` exists. If it doesn't:

1. Ask the user for their Rigor Mining server URL (e.g.
   `https://your-server.example.com` or `http://localhost:3000`) and an
   API token. Tell them the token is created on the app's Settings page
   (`/settings`) if they don't already have one.
2. Run:
   ```
   python3 scripts/rigormining.py configure --server "<server>" --token "<token>"
   ```
3. This validates the token against `GET /api/v1/whoami` before saving -
   if it fails, the command exits with the server's error; ask the user
   to double check the URL/token and try again.

If the user ever wants to switch servers or rotate the token, just re-run
`configure` - it overwrites the existing config.

## 2. Find the Kobo

Run:
```
python3 scripts/rigormining.py find-kobo
```
This checks common mount locations (`/Volumes/*` on macOS,
`/media/$USER/*` and `/run/media/$USER/*` on Linux) for a directory
containing a `.kobo` folder, which every real Kobo has at its root. If it
prints a path, that's `--kobo-dir` for step 4. If it exits with nothing,
ask the user directly where the Kobo is mounted (on Windows this will
always be the case - ask for the drive letter, e.g. `E:\`) and use that
instead.

## 3. If the user names specific items ("sync the X paper")

Don't default to syncing everything queued without being asked - someone
naming specific items usually wants just those, not a bulk sync. Search
first:
```
python3 scripts/rigormining.py search "<query>"
```
Each line is `id`, current `sync_state`, and `title`. If there's exactly
one clear match, queue it:
```
python3 scripts/rigormining.py queue <item_id>
```
If there are several plausible matches, show them and ask which one they
meant before queuing anything. To cancel/remove something instead:
```
python3 scripts/rigormining.py unqueue <item_id>
```

If the user just says "sync my queued items" / "sync the pdfs" with no
specifics, skip this step - step 4 processes whatever's already queued
(from the app's UI or a prior `queue` call), which is exactly the
"only the one or two I actually picked" behavior the Sync screen's
checkboxes already give (they default to unchecked).

## 4. Run the sync

```
python3 scripts/rigormining.py run --kobo-dir "<path from step 2>"
```

This, in order:
1. Reconciles: any item the server thinks is `synced` but whose file is
   actually missing from the device (deleted directly on the Kobo) gets
   un-queued.
2. Processes `request_remove` items: deletes the file from the device,
   then acks the server.
3. Processes `request_sync` items: downloads the file from the server,
   writes it into a `rigormining` folder on the Kobo, then acks the
   server.

It prints per-item progress and a final summary line. Report that summary
back to the user in your own words - don't just paste the raw output.
Add `--dry-run` first if the user wants to preview without touching
anything.

## Adding papers ("send X to my Rigor Mining server")

Needs only the config from step 1, no Kobo. Pass an arXiv id or any
paper URL:
```
python3 scripts/rigormining.py add 2112.04426
python3 scripts/rigormining.py add https://arxiv.org/abs/2306.11644 --queue
python3 scripts/rigormining.py add https://aclanthology.org/N19-1423/
```
This posts to `/api/v1/capture` exactly like the browser extension: it
reads the page's `citation_*` meta tags (title, authors, date, venue,
DOI) so the server builds the citation, and downloads the PDF from
`citation_pdf_url` (arXiv's PDF is found directly). A direct PDF URL
works too but comes in with no citation metadata; `--file paper.pdf`
attaches a local PDF instead of downloading one.

- If the user names a paper rather than giving an id/URL, work out the
  arXiv id (or the publisher's page) first, and say which one you used.
- It refuses when an item with the same title already exists; pass
  `--force` only if the user really wants a second copy.
- `--dry-run` shows the title, author count and PDF size without adding
  anything; use it when you're unsure the URL is the right paper.
- `--queue` also queues it for the Kobo (same as step 3's `queue`).
