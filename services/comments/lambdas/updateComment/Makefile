build:
	@echo 'Building update comment lambda...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping update comment...'
	zip -j main.zip main

