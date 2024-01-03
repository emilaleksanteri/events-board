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

## synth: synth local cdk
.PHONY: synth
synth:
	npx cdklocal synth

## deploy: deploy local cdk
.PHONY: deploy
deploy:
	npx cdklocal deploy --all

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

## build/lambdas: build all apis
.PHONY: build/lambdas
build/lambdas:
	@echo "Building go app"
	cd ./services/posts && make build/lambdas
	cd ./services/comments && make build/lambdas
## tidy/lambdas: go mod tidy for all lambdas
.PHONY: tidy/lambdas
tidy/lambdas:
	@echo "Tidying app modules"
	cd ./services/posts && make tidy/lambdas
	cd ./services/comments && make tidy/lambdas
