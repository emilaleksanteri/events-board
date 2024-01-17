build:
	@echo 'Building get likes lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping get likes...'
	zip -j main.zip main
