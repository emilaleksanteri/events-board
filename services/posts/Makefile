export AWS_ACCESS_KEY_ID ?= test
export AWS_SECRET_ACCESS_KEY ?= test
export AWS_DEFAULT_REGION=us-east-1
export LAMBDA_RUNTIME_ENVIRONMENT_TIMEOUT=300
VENV_DIR ?= .venv

bootstrap:
	npx cdklocal bootstrap

deploy:
	npx cdklocal deploy

start:
	docker-compose up
stop:
	localstack stop
