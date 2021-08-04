PHONY: clean build

clean:
	rm -rf dist

build:
	yarn build

start:
	yarn start

publish:
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		dist s3://knotset.com/
	aws cloudfront create-invalidation \
		--distribution-id E3CH03HZXMU8EY \
		--paths "/*"

