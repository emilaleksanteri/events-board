include .envrc
build:
	@echo 'Building posts...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main main.go

zip:
	@echo 'Zipping posts...'
	zip -j main.zip main

test/aws:
	@make build
	@make zip
	docker build --platform linux/amd64 -t docker-image:test .

test/aws/run:
	docker run --platform linux/amd64 -p 9000:8080 docker-image:test

test/aws/call:
	curl "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}'
