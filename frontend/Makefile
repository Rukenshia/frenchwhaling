CLOUDFRONT_ID=E70E85S3UWCMA

.PHONY: build
build:
	npm run build
	rm public/bundle.css.map

.PHONY: deploy
deploy:
	aws s3 sync public s3://whaling.in.fkn.space
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID) --paths /index.html /img/* /bundle.css /bundle.js /global.css /warships.min.json
