build:
	@echo 'Building post like lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping post like...'
	zip -j main.zip main
