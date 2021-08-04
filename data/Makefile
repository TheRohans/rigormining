PHONY: build public clean publish

publish:
	aws s3 sync --delete --region us-west-2 \
		--cache-control max-age=604800 \
		public s3://private.knotset.com/
