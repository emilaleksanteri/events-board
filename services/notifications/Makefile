.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build/lambdas: build all lambdas for api
.PHONY: build/lambdas
build/lambdas:
	@echo "Building go app"
	cd ./lambdas/connectionHandler && make build && make zip
	cd ./lambdas/eventBusMsgHandler && make build && make zip
	cd ./lambdas/messageHandler && make build && make zip

## tidy/lambdas: go mod tidy for all lambdas
.PHONY: tidy/lambdas
tidy/lambdas:
	@echo "Tidying app modules"
	cd ./lambdas/connectionHandler && go mod tidy
	cd ./lambdas/eventBusMsgHandler && go mod tidy
	cd ./lambdas/messageHandler && go mod tidy
