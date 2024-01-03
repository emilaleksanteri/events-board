.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build/lambdas: build all lambdas for api
.PHONY: build/lambdas
build/lambdas:
	@echo "Building go app"
	cd ./lambdas/postComment && make build && make zip
	cd ./lambdas/getComment && make build && make zip
	cd ./lambdas/updateComment && make build && make zip

## tidy/lambdas: go mod tidy for all lambdas
.PHONY: tidy/lambdas
tidy/lambdas:
	@echo "Tidying app modules"
	cd ./lambdas/postComment && go mod tidy
	cd ./lambdas/getComment && go mod tidy
	cd ./lambdas/updateComment && go mod tidy
