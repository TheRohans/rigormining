PHONY: build public clean publish

build:
	rm ./public/public/library/rohan/metadata.json
	cd ./public/public/library/rohan; \
		node toMeta.js . > metadata.json
	cat ./public/public/library/rohan/metadata.json

publish:
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		--acl bucket-owner-full-control \
		public s3://private.knotset.com/
