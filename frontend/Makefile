PHONY: clean build

hash = $(shell git log --pretty=format:'%h' -n 1)

RIGORMINING_VERSION=development

install:
	npm install
	npx browserslist@latest --update-db
	$(MAKE) copy_worker

clean:
	rm -rf dist
	rm -rf coverage

make_dev_env:
	@echo "" > .env
	@echo "RIGORMINING_BUCKET=http://localhost:8000" >> .env
	@echo "RIGORMINING_VERSION=$(hash)" >> .env
	@echo "RIGORMINING_API_BASE=http://localhost:3000" >> .env
	@echo "" >> .env

make_prod_env:
	@echo "" > .env
	@echo "RIGORMINING_BUCKET=https://rigormining.com" >> .env
	@echo "RIGORMINING_VERSION=$(hash)" >> .env
	@echo "" >> .env

copy_worker:
	cp ./node_modules/pdfjs-dist/build/pdf.worker.js ./pdf.worker.js

# Build for distro
build: clean make_prod_env copy_worker
	npm run build
	cp pdf.worker.js dist/pdf.worker.js

# Build for dev
start: make_dev_env copy_worker
	npm start

test:
	npm test

coverage:
	npm run test:coverage

# Deploy
# NOTE: predates the Go backend serving frontend/dist directly (see repo
# root README) - this S3/CloudFront path is probably superseded, not
# something actively used anymore.
publish: build
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		dist s3://rigormining.com/
	aws cloudfront create-invalidation \
		--distribution-id E3CH03HZXMU8EY \
		--paths "/*"
