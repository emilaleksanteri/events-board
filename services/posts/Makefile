include .envrc
build:
	@echo 'Building posts...'
	go build -ldflags='-s' -o=./bin/posts
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/posts 

zip:
	@echo 'Zipping posts...'
	zip -j ./bin/posts.zip ./bin/posts

run:
	@make build
	@echo 'Running posts...'
	./bin/posts -db-username=${DB_USERNAME} -db-host=${DB_HOST} -db-name=${DB_NAME} -db-password=${DB_PASSWORD} -db-port=${DB_PORT} -db-region=${DB_REGION}

dev:
	@echo 'Running posts dev...'
	go run main.go -db-username=${DB_USERNAME} -db-host=${DB_HOST} -db-name=${DB_NAME} -db-password=${DB_PASSWORD} -db-port=${DB_PORT} -db-region=${DB_REGION}

localstack:
	@echo 'Running posts localstack...'
	ACTIVATE_PRO=0 localstack start

dev/lambda:
	@echo 'Turning posts into a lambda function'
	@make build
	@make zip
	awslocal lambda create-function \
		--function-name posts \
		--runtime go1.x \
		--handler lambda.apiHandler \
		--memory-size 128 \
		--zip-file fileb://bin/posts.zip \
		--role arn:aws:iam::111111111111:role/apigw
