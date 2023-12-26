include .envrc
build:
	@echo 'Building posts...'
	go build -ldflags='-s' -o=./bin/posts
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/posts 

run:
	@make build
	@echo 'Running posts...'
	./bin/posts -db-username=${DB_USERNAME} -db-host=${DB_HOST} -db-name=${DB_NAME} -db-password=${DB_PASSWORD} -db-port=${DB_PORT} -db-region=${DB_REGION}

dev:
	@echo 'Running posts dev...'
	go run main.go -db-username=${DB_USERNAME} -db-host=${DB_HOST} -db-name=${DB_NAME} -db-password=${DB_PASSWORD} -db-port=${DB_PORT} -db-region=${DB_REGION}
