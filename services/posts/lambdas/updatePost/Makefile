build:
	@echo 'Building update post lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping update posts...'
	zip -j main.zip main
