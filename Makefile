PHONY: clean build

hash = $(shell git log --pretty=format:'%h' -n 1)

clean:
	rm -rf dist

copy_worker:
	cp ./node_modules/pdfjs-dist/build/pdf.worker.js ./pdf.worker.js

build: clean
	yarn build --env KNOTSET_VERSION=$(hash)
	cp pdf.worker.js dist/pdf.worker.js

start:
	yarn start --env KNOTSET_VERSION=$(hash)

publish: build
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		dist s3://knotset.com/
	aws cloudfront create-invalidation \
		--distribution-id E3CH03HZXMU8EY \
		--paths "/*"
