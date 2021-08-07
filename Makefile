PHONY: clean build

hash = $(shell git log --pretty=format:'%h' -n 1)

clean:
	rm -rf dist

build: clean
	yarn build --env KNOTSET_VERSION=$(hash)

start:
	yarn start --env KNOTSET_VERSION=$(hash)

publish: build
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		dist s3://knotset.com/
	aws cloudfront create-invalidation \
		--distribution-id E3CH03HZXMU8EY \
		--paths "/*"

