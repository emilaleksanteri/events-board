build:
	@echo 'Building delete comment lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping delete comment...'
	zip -j main.zip main
