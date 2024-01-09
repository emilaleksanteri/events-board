include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage: '
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'


.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]


.PHONY: db/psql
db/psql:
	psql ${DB}

## db/migrations/up: apply all up db migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DB} up

## db/migrations/new name=$1: create a new db migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/force version=$1: force a db migration version
.PHONY: db/migrations/force
db/migrations/force:
	@echo 'Forcing migration version ${version}'
	migrate -path ./migrations -database=${DB} force ${version}
