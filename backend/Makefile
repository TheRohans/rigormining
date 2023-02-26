PHONY: build public clean publish serve

include .env
export

install: 
	go mod tidy
	cp .env.template .env

start: build_index
	go run cmd/server/main.go

build_index:
	rm ./public/public/library/rohan/metadata.json
	cd ./public/public/library/rohan; \
		node toMeta.js . > metadata.json
	cat ./public/public/library/rohan/metadata.json

# publish: build
# 	aws s3 sync --delete --region us-west-2 \
# 		--cache-control max-age=604800 \
# 		--acl bucket-owner-full-control \
# 		public s3://private.knotset.com/
