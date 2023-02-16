PHONY: clean build

hash = $(shell git log --pretty=format:'%h' -n 1)

KNOTSET_VERSION=development

install:
	yarn install
	npx browserslist@latest --update-db

clean:
	rm -rf dist
	rm -rf coverage

make_env:
	@echo "" > .env
	@echo "KNOTSET_VERSION=$(hash)" >> .env
	@echo "" >> .env

copy_worker:
	cp ./node_modules/pdfjs-dist/build/pdf.worker.js ./pdf.worker.js

# Build for distro
build: clean make_env copy_worker
	NODE_OPTIONS=--openssl-legacy-provider yarn build
	cp pdf.worker.js dist/pdf.worker.js

# Build for dev
start: make_env
	NODE_OPTIONS=--openssl-legacy-provider yarn start

test:
	yarn test

coverage:
	yarn test:coverage

# Deploy
publish: build
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		dist s3://knotset.com/
	aws cloudfront create-invalidation \
		--distribution-id E3CH03HZXMU8EY \
		--paths "/*"
