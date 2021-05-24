.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o trainingstatbot -v cmd/trainingstatbot/main.go