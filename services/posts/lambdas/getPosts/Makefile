build:
	@echo 'Building get posts lambdas...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping get posts...'
	zip -j main.zip main
