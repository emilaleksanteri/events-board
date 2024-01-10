build:
	@echo 'Building post posts lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping post posts...'
	zip -j main.zip main
