.PHONY: deps
deps:
	GOPRIVATE="github.com/QualiArts" go mod download

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o bin/octo-cli-linux-amd64 ./src/octo-cli
	GOOS=windows GOARCH=amd64 go build -o bin/octo-cli-windows-amd64.exe ./src/octo-cli
	GOOS=darwin GOARCH=amd64 go build -o bin/octo-cli-darwin-amd64 ./src/octo-cli
