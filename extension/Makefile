.PHONY: build clean

# Same source works unpacked in both browsers today (manifest.json ships
# both the MV3 service_worker key Chrome wants and the scripts fallback
# array Firefox wants) - two names now just gives each store/beta channel
# its own artifact to diverge from later without restructuring anything.
SOURCES = background.js content.js manifest.json options.html options.js popup.html popup.js

build: clean
	mkdir -p dist
	zip -j dist/rigormining-chrome.zip $(SOURCES)
	zip -j dist/rigormining-firefox.zip $(SOURCES)

clean:
	rm -rf dist
