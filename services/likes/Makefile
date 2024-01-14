.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build/lambdas: build all lambdas for api
.PHONY: build/lambdas
build/lambdas:
	@echo "Building likes go app"
	cd ./lambdas/postLike && make build && make zip

## tidy/lambdas: go mod tidy for all lambdas
.PHONY: tidy/lambdas
tidy/lambdas:
	@echo "Tidying app modules"
	cd ./lambdas/postLike && go mod tidy
