build:
	@echo 'Building connection handler for ws notifications...'
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build  -o main

zip:
	@echo 'Zipping posts...'
	zip -j main.zip main
