build:
	@echo 'Building get comment(s) lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping get comment(s)...'
	zip -j main.zip main
