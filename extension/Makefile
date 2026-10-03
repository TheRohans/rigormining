.PHONY: build clean version icons release_firefox

# Same source works unpacked in both browsers today (manifest.json ships
# both the MV3 service_worker key Chrome wants and the scripts fallback
# array Firefox wants) - two names now just gives each store/beta channel
# its own artifact to diverge from later without restructuring anything.
SOURCES = background.js content.js manifest.json options.html options.js popup.html popup.js icons

# Comes from the root Makefile (see VERSION there); manifest.json's own
# version is just a placeholder that gets overwritten in the zips. Docker
# has no root Makefile in this stage, so it passes VERSION in.
VERSION ?= $(shell $(MAKE) -s --no-print-directory -C .. version)

version:
	@echo $(VERSION)

build: clean
	mkdir -p dist/stage
	cp -R $(SOURCES) dist/stage/
	sed 's/"version": *"[^"]*"/"version": "$(VERSION)"/' manifest.json > dist/stage/manifest.json
	cd dist/stage && zip -r ../rigormining-chrome.zip $(SOURCES)
	cd dist/stage && zip -r ../rigormining-firefox.zip $(SOURCES)
	rm -rf dist/stage
	cp signed/* dist/ 2>/dev/null || true

clean:
	rm -rf dist

# The PNGs are committed so build (and Docker) don't need librsvg; rerun
# this after editing icon.svg.
icons:
	for s in 16 32 48 96 128; do rsvg-convert -w $$s -h $$s icon.svg -o icons/icon-$$s.png; done

# Firefox self-distribution: upload dist/rigormining-firefox.zip to AMO as
# "On your own", download the signed .xpi it gives back, then
#   make release_firefox XPI=~/Downloads/<signed>.xpi
# and commit signed/. Installed copies poll updates.json (the update_url in
# manifest.json) and pull the .xpi from the backend's public
# /extension/firefox/ route. Only the latest version needs to be listed.
UPDATE_BASE = https://rigormining.therohans.com/extension/firefox
release_firefox:
	@test -n "$(XPI)" || { echo "usage: make release_firefox XPI=path/to/signed.xpi"; exit 1; }
	mkdir -p signed
	cp "$(XPI)" signed/rigormining-firefox.xpi
	v=$$(unzip -p signed/rigormining-firefox.xpi manifest.json | sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' | head -1); \
	test -n "$$v" || { echo "couldn't read version from $(XPI)"; exit 1; }; \
	printf '{\n  "addons": {\n    "rigormining@therohans.com": {\n      "updates": [\n        {\n          "version": "%s",\n          "update_link": "$(UPDATE_BASE)/rigormining-firefox.xpi"\n        }\n      ]\n    }\n  }\n}\n' "$$v" > signed/updates.json; \
	echo "signed/ now advertises $$v"
