PHONY: build public clean publish serve

install: 
	yarn add http-server

start: build_index
	npx http-server -p 8000 --cors

build_index:
	rm ./public/public/library/rohan/metadata.json
	cd ./public/public/library/rohan; \
		node toMeta.js . > metadata.json
	cat ./public/public/library/rohan/metadata.json

publish: build
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		--acl bucket-owner-full-control \
		public s3://private.knotset.com/
