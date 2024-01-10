.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build/lambdas: build all lambdas for api
.PHONY: build/lambdas
build/lambdas:
	@echo "Building posts go app"
	cd ./lambdas/getPosts && make build && make zip
	cd ./lambdas/postPost && make build && make zip
	cd ./lambdas/updatePost && make build && make zip
	cd ./lambdas/deletePost && make build && make zip

## tidy/lambdas: go mod tidy for all lambdas
.PHONY: tidy/lambdas
tidy/lambdas:
	@echo "Tidying app modules"
	cd ./lambdas/getPosts && go mod tidy
	cd ./lambdas/postPost && go mod tidy
	cd ./lambdas/updatePost && go mod tidy
	cd ./lambdas/deletePost && go mod tidy
