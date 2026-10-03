.PHONY: build clean version

# Same source works unpacked in both browsers today (manifest.json ships
# both the MV3 service_worker key Chrome wants and the scripts fallback
# array Firefox wants) - two names now just gives each store/beta channel
# its own artifact to diverge from later without restructuring anything.
SOURCES = background.js content.js manifest.json options.html options.js popup.html popup.js

# Comes from the root Makefile (see VERSION there); manifest.json's own
# version is just a placeholder that gets overwritten in the zips. Docker
# has no root Makefile in this stage, so it passes VERSION in.
VERSION ?= $(shell $(MAKE) -s --no-print-directory -C .. version)

version:
	@echo $(VERSION)

build: clean
	mkdir -p dist/stage
	cp $(SOURCES) dist/stage/
	sed 's/"version": *"[^"]*"/"version": "$(VERSION)"/' manifest.json > dist/stage/manifest.json
	cd dist/stage && zip ../rigormining-chrome.zip $(SOURCES)
	cd dist/stage && zip ../rigormining-firefox.zip $(SOURCES)
	rm -rf dist/stage

clean:
	rm -rf dist
