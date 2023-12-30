include .envrc
export AWS_ACCESS_KEY_ID ?= test
export AWS_SECRET_ACCESS_KEY ?= test
export AWS_DEFAULT_REGION=us-east-1
export LAMBDA_RUNTIME_ENVIRONMENT_TIMEOUT=300
VENV_DIR ?= .venv

help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

bootstrap:
	npx cdklocal bootstrap

deploy:
	npx cdklocal deploy

start:
	docker-compose up -d

stop:
	docker-compose down

build/infra:
	@echo "Building infra"
	npm run build

build/app:
	@echo "Building go app"
	cd ./app && make build && make zip
