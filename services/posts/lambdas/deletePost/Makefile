build:
	@echo 'Building delete posts lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping delete post...'
	zip -j main.zip main
