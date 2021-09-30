PHONY: clean build

hash = $(shell git log --pretty=format:'%h' -n 1)

KNOTSET_identityPoolId=us-west-2:4800e468-c8b7-4a0b-a772-6add33659941
KNOTSET_VERSION=development

clean:
	rm -rf dist

make_env:
	@echo "" > .env
	@echo "KNOTSET_identityPoolId=$(KNOTSET_identityPoolId)" >> .env
	@echo "KNOTSET_VERSION=$(hash)" >> .env
	@echo "" >> .env

copy_worker:
	cp ./node_modules/pdfjs-dist/build/pdf.worker.js ./pdf.worker.js

build: clean make_env copy_worker
	yarn build
	cp pdf.worker.js dist/pdf.worker.js

start: make_env
	yarn start

publish: build
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		dist s3://knotset.com/
	aws cloudfront create-invalidation \
		--distribution-id E3CH03HZXMU8EY \
		--paths "/*"

# Some next steps:
# "amplify status" will show you what you've added already and if it's locally configured or deployed
# "amplify add <category>" will allow you to add features like user login or a backend API
# "amplify push" will build all your local backend resources and provision it in the cloud
# "amplify console" to open the Amplify Console and view your project status
# "amplify publish" will build all your local backend and frontend resources (if you have hosting category added) and provision it in the cloud

# Pro tip:
# Try "amplify add api" to create a backend API and then "amplify publish" to deploy everything

# GraphQL endpoint: https://3lcceo53enhcxnxddvoaufnd4u.appsync-api.us-west-2.amazonaws.com/graphql
