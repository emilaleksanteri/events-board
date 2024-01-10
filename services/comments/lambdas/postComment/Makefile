build:
	@echo 'Building post comment lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping post comment...'
	zip -j main.zip main

