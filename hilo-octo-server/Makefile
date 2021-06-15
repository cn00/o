REV=$(shell git rev-parse --short HEAD)

# echo "VERSION:$(VERSION) REV:$(REV)"
all: api admin cli
api:
	CGO_ENABLED=0 go build -x -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-api
apirun:
	CGO_ENABLED=0 go run -x -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-api
admin:
	CGO_ENABLED=0 go build -x -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-admin
admrun:
	CGO_ENABLED=0 go run -x -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-admin
cli:
	CGO_ENABLED=0 go build -x -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-cli
clirun:
	CGO_ENABLED=0 go run -x -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-cli
