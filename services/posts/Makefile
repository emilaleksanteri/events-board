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

## build/app: builds go app
.PHONY: build/app
build/app:
	@echo "Building go app"
	cd ./app && make build && make zip

## tidy/app: go mod tidy for app dir
.PHONY: tidy/app
tidy/app:
	@echo "Tidying app modules"
	cd ./app && go mod tidy
