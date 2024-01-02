include .envrc
export AWS_ACCESS_KEY_ID ?= test
export AWS_SECRET_ACCESS_KEY ?= test
export AWS_DEFAULT_REGION=us-east-1
export LAMBDA_RUNTIME_ENVIRONMENT_TIMEOUT=300
VENV_DIR ?= .venv


.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## bootstrap: bootstrap cdk
.PHONY: bootstrap
bootstrap:
	npx cdklocal bootstrap

## deploy: deploy local cdk
.PHONY: deploy
deploy:
	npx cdklocal deploy

## start: start localstack
.PHONY: start
start:
	docker-compose up -d

## stop: stop localstack
.PHONY: stop
stop:
	docker-compose down

## build/infra: builds cdk infra stuff
.PHONY: build/infra
build/infra:
	@echo "Building infra"
	npm run build

## build/lambdas: build all lambdas for api
.PHONY: build/lambdas
build/lambdas:
	@echo "Building go app"
	cd ./lambdas/getPosts && make build && make zip
	cd ./lambdas/postPost && make build && make zip
	cd ./lambdas/updatePost && make build && make zip

## tidy/lambdas: go mod tidy for all lambdas
.PHONY: tidy/lambdas
tidy/lambdas:
	@echo "Tidying app modules"
	cd ./lambdas/getPosts && go mod tidy
	cd ./lambdas/postPost && go mod tidy
	cd ./lambdas/updatePost && go mod tidy
