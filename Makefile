.PHONY: test

build-ui:
	cd ./internal/dashboard/app && npm run build

release:
	git tag broker/xkafka/v0.0.1-alpha.2
	git tag broker/xrabbitmq/v0.0.1-alpha.2
	git tag storage/xgorm/v0.0.1-alpha.2
	git tag storage/xmysql/v0.0.1-alpha.2
