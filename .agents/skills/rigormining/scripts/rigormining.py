#!/usr/bin/env python3
"""
Kobo sync helper for a Rigor Mining server.

Talks to a running Rigor Mining server's REST API (see the repo's
README.md and backend/cmd/server/main.go for the routes) to copy items
queued with sync_state=request_sync onto a locally mounted Kobo, delete
items queued with sync_state=request_remove, and report back what
happened. It can also add papers to the library (arXiv ids or URLs),
the same way the browser extension's capture does. Config (server URL + API token) lives in a local file outside
this skill and outside the repo - nothing in this script is tied to any
particular server, account, or machine; every path is discovered or
passed in at runtime.

Uses only the Python standard library - no extra dependencies to install.
"""
import argparse
import html
import json
import os
import re
import stat
import sys
import urllib.error
import urllib.parse
import urllib.request
import uuid
from pathlib import Path

DEFAULT_CONFIG = Path.home() / ".rigormining-sync.json"
# Matches the folder name the browser's Sync-to-Kobo screen already uses
# (frontend/src/components/Sync/index.tsx) - keeping it the same means
# either tool can reconcile against files the other one put there.
SUBDIR_NAME = "rigormining"


def load_config(path: Path) -> dict:
    if not path.exists():
        sys.exit(f"No config at {path} yet - run the 'configure' command first.")
    return json.loads(path.read_text())


def save_config(path: Path, server: str, token: str) -> None:
    path.write_text(json.dumps({"server": server.rstrip("/"), "token": token}))
    os.chmod(path, stat.S_IRUSR | stat.S_IWUSR)  # 0600 - owner read/write only


def api_request(cfg: dict, method: str, path: str, body: dict = None, raw: bool = False):
    url = f"{cfg['server']}{path}"
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Authorization", f"Bearer {cfg['token']}")
    if data is not None:
        req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            payload = resp.read()
            return payload if raw else (json.loads(payload) if payload else None)
    except urllib.error.HTTPError as e:
        detail = e.read().decode(errors="replace")
        sys.exit(f"{method} {path} failed: {e.code} {detail}")
    except urllib.error.URLError as e:
        sys.exit(f"{method} {path} failed: {e.reason}")


def sanitize(name: str) -> str:
    # Mirrors frontend/src/components/Sync/index.tsx's sanitize() exactly,
    # so filenames match whichever tool (browser or this skill) wrote them.
    return re.sub(r"[\\/:]", "-", name)


def cmd_configure(args):
    cfg = {"server": args.server.rstrip("/"), "token": args.token}
    who = api_request(cfg, "GET", "/api/v1/whoami")
    save_config(Path(args.config), args.server, args.token)
    print(f"Saved config to {args.config} - signed in as {who.get('email', '(unknown)')}.")


def cmd_find_kobo(_args):
    candidates = []
    if sys.platform == "darwin":
        volumes = Path("/Volumes")
        if volumes.is_dir():
            candidates.extend(volumes.iterdir())
    else:
        user = os.environ.get("USER", "")
        for base in (f"/media/{user}", f"/run/media/{user}", "/media"):
            p = Path(base)
            if p.is_dir():
                candidates.extend(p.iterdir())

    # Every real Kobo has a .kobo folder at the root of its filesystem -
    # more reliable than matching on the "KOBOeReader" volume label, which
    # a user could have renamed.
    for c in candidates:
        if (c / ".kobo").is_dir():
            print(str(c))
            return
    sys.exit("No mounted Kobo found automatically - pass its path directly with --kobo-dir.")


def cmd_search(args):
    cfg = load_config(Path(args.config))
    q = urllib.parse.quote(args.query)
    items = api_request(cfg, "GET", f"/api/v1/items?q={q}")
    if not items:
        print("No matches.")
        return
    for i in items:
        print(f"{i['id']}\t{i.get('sync_state') or 'not queued'}\t{i['title']}")


def cmd_queue(args):
    cfg = load_config(Path(args.config))
    item = api_request(cfg, "POST", f"/api/v1/items/{args.item_id}/sync")
    print(f"Queued: {item['title']} ({item.get('sync_state')})")


def cmd_unqueue(args):
    cfg = load_config(Path(args.config))
    item = api_request(cfg, "DELETE", f"/api/v1/items/{args.item_id}/sync")
    print(f"Un-queued: {item['title']} ({item.get('sync_state') or 'not queued'})")


ARXIV_ID = re.compile(r"(?:arxiv\.org/(?:abs|pdf)/|^arxiv:)?(\d{4}\.\d{4,5}(?:v\d+)?)(?:\.pdf)?/?$", re.I)
META_TAG = re.compile(r"<meta\s+[^>]*>", re.I)


def fetch(url: str) -> tuple:
    req = urllib.request.Request(url, headers={"User-Agent": "rigormining-skill/1.0"})
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            return resp.read(), resp.headers.get_content_type()
    except (urllib.error.HTTPError, urllib.error.URLError) as e:
        sys.exit(f"GET {url} failed: {e}")


def meta_attr(tag: str, attr: str):
    # "double", 'single' or bare values - all valid HTML (ACL Anthology uses bare names)
    m = re.search(rf"""\b{attr}\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+))""", tag, re.I)
    return next((g for g in m.groups() if g is not None), None) if m else None


def page_meta(page: str) -> list:
    # The same {name, content} list the browser extension sends: every
    # <meta name=... content=...> on the page, in either attribute order.
    # The server's capture handler picks the citation_* / dc.* ones out.
    tags = []
    for tag in META_TAG.findall(page):
        name, content = meta_attr(tag, "name"), meta_attr(tag, "content")
        if name is not None and content is not None:
            tags.append({"name": name, "content": html.unescape(content)})
    return tags


def multipart(fields: dict, file_bytes: bytes = None, filename: str = "capture.pdf") -> tuple:
    boundary = uuid.uuid4().hex
    parts = [f'--{boundary}\r\nContent-Disposition: form-data; name="{k}"\r\n\r\n{v}\r\n'.encode()
             for k, v in fields.items()]
    if file_bytes is not None:
        parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{filename}"\r\n'
                     f"Content-Type: application/pdf\r\n\r\n".encode() + file_bytes + b"\r\n")
    return b"".join(parts) + f"--{boundary}--\r\n".encode(), f"multipart/form-data; boundary={boundary}"


def cmd_add(args):
    # Adds a paper the same way the browser extension's capture does
    # (POST /api/v1/capture with url, title, meta and the PDF), so the
    # server does the citation extraction exactly as it would for a click.
    cfg = load_config(Path(args.config))
    m = ARXIV_ID.search(args.source.strip())
    if m:  # arXiv: the abs page has full citation_* tags, the pdf is predictable
        page_url, pdf_url = f"https://arxiv.org/abs/{m.group(1)}", f"https://arxiv.org/pdf/{m.group(1)}"
    else:
        page_url, pdf_url = args.source.strip(), None

    body, ctype = fetch(page_url)
    if ctype == "application/pdf":  # a direct PDF link, no page to read tags from
        meta, pdf = [], body
        title = Path(urllib.parse.urlparse(page_url).path).stem or page_url
    else:
        meta = page_meta(body.decode("utf-8", "replace"))
        get = lambda n: next((t["content"] for t in meta if t["name"].lower() == n and t["content"]), None)
        title_tag = re.search(r"<title[^>]*>(.*?)</title>", body.decode("utf-8", "replace"), re.I | re.S)
        title = get("citation_title") or get("dc.title") or (html.unescape(title_tag.group(1)).strip() if title_tag else page_url)
        pdf_url = pdf_url or get("citation_pdf_url")
        pdf = None
        if args.file:
            pdf = Path(args.file).read_bytes()
        elif pdf_url:
            pdf, _ = fetch(pdf_url)
    if pdf is not None and pdf[:4] != b"%PDF":
        print(f"Warning: {pdf_url or args.file} is not a PDF - adding without a file.")
        pdf = None

    if not args.force:
        q = urllib.parse.quote(title)
        for i in api_request(cfg, "GET", f"/api/v1/items?q={q}") or []:
            if i["title"].strip().lower() == title.strip().lower():
                sys.exit(f"Already in the library: {i['id']}\t{i['title']} (use --force to add anyway)")

    if args.dry_run:
        authors = sum(t["name"].lower() == "citation_author" for t in meta)
        print(f"Would add: {title}\n  url {page_url}\n  {len(meta)} meta tags, {authors} authors, "
              f"{f'{len(pdf) // 1024} KB PDF' if pdf else 'no file'}")
        return
    data, content_type = multipart({"url": page_url, "title": title, "meta": json.dumps(meta)}, pdf)
    req = urllib.request.Request(f"{cfg['server']}/api/v1/capture", data=data, method="POST")
    req.add_header("Authorization", f"Bearer {cfg['token']}")
    req.add_header("Content-Type", content_type)
    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            item = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        sys.exit(f"POST /api/v1/capture failed: {e.code} {e.read().decode(errors='replace')}")
    except urllib.error.URLError as e:
        sys.exit(f"POST /api/v1/capture failed: {e.reason}")
    size = f"{len(pdf) // 1024} KB PDF" if pdf else "no file"
    print(f"{item['id']}\tadded\t{item['title']} ({size})")
    if args.queue:
        cmd_queue(argparse.Namespace(config=args.config, item_id=item["id"]))


def cmd_run(args):
    cfg = load_config(Path(args.config))
    mount = Path(args.kobo_dir)
    if not mount.is_dir():
        sys.exit(f"{mount} does not exist - is the Kobo actually mounted there?")
    kobo_dir = mount / SUBDIR_NAME
    kobo_dir.mkdir(parents=True, exist_ok=True)

    # Reconcile first: sync_state=="synced" only ever records "we copied
    # this at some point" - if the user deleted the file directly on the
    # device, nothing else would ever notice. Same check the browser's
    # Sync screen does before every run.
    synced = api_request(cfg, "GET", "/api/v1/items?sync_state=synced") or []
    reconciled = 0
    for item in synced:
        if not item.get("file_type"):
            continue
        filename = sanitize(f"{item['title']}.{item['file_type']}")
        if not (kobo_dir / filename).exists():
            api_request(cfg, "POST", f"/api/v1/items/{item['id']}/sync/ack", {"result": "removed"})
            reconciled += 1
    if reconciled:
        print(f"{reconciled} item(s) no longer on the device - un-queued.")

    to_remove = api_request(cfg, "GET", "/api/v1/items?sync_state=request_remove") or []
    removed = 0
    for item in to_remove:
        filename = sanitize(f"{item['title']}.{item.get('file_type') or ''}")
        target = kobo_dir / filename
        if args.dry_run:
            print(f"[dry-run] would remove {filename}")
            continue
        if target.exists():
            target.unlink()
        api_request(cfg, "POST", f"/api/v1/items/{item['id']}/sync/ack", {"result": "removed"})
        removed += 1
        print(f"Removed: {item['title']}")

    to_sync = api_request(cfg, "GET", "/api/v1/items?sync_state=request_sync") or []
    copied = 0
    for item in to_sync:
        if not item.get("file_type"):
            print(f"Skipping (no file attached yet): {item['title']}")
            continue
        filename = sanitize(f"{item['title']}.{item['file_type']}")
        if args.dry_run:
            print(f"[dry-run] would copy {filename}")
            continue
        data = api_request(cfg, "GET", f"/api/v1/items/{item['id']}/file", raw=True)
        (kobo_dir / filename).write_bytes(data)
        api_request(cfg, "POST", f"/api/v1/items/{item['id']}/sync/ack", {"result": "synced"})
        copied += 1
        print(f"Copied: {item['title']}")

    print(f"\nDone - {copied} copied, {removed} removed, {reconciled} un-queued (already gone from the device).")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", default=str(DEFAULT_CONFIG), help="path to the local config file")
    sub = parser.add_subparsers(dest="command", required=True)

    p = sub.add_parser("configure", help="save the server URL + API token")
    p.add_argument("--server", required=True)
    p.add_argument("--token", required=True)
    p.set_defaults(func=cmd_configure)

    p = sub.add_parser("find-kobo", help="best-effort autodetect of a mounted Kobo")
    p.set_defaults(func=cmd_find_kobo)

    p = sub.add_parser("search", help="find items by title/author")
    p.add_argument("query")
    p.set_defaults(func=cmd_search)

    p = sub.add_parser("queue", help="queue one item for sync (sets request_sync)")
    p.add_argument("item_id")
    p.set_defaults(func=cmd_queue)

    p = sub.add_parser("unqueue", help="un-queue one item (cancels, or requests removal if already on-device)")
    p.add_argument("item_id")
    p.set_defaults(func=cmd_unqueue)

    p = sub.add_parser("add", help="add a paper by arXiv id / URL (or any page or PDF URL)")
    p.add_argument("source", help="e.g. 2112.04426, https://arxiv.org/abs/2112.04426, or a paper's page/PDF URL")
    p.add_argument("--file", help="attach this local PDF instead of downloading one")
    p.add_argument("--queue", action="store_true", help="also queue it for the Kobo")
    p.add_argument("--force", action="store_true", help="add even if an item with the same title exists")
    p.add_argument("--dry-run", action="store_true", help="show what would be added, add nothing")
    p.set_defaults(func=cmd_add)

    p = sub.add_parser("run", help="process whatever's queued: copy/remove/ack")
    p.add_argument("--kobo-dir", required=True, help="the mounted Kobo's root, e.g. /Volumes/KOBOeReader")
    p.add_argument("--dry-run", action="store_true")
    p.set_defaults(func=cmd_run)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
