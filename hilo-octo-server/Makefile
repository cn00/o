REV=$(shell git rev-parse --short HEAD)
all: api admin
api:
	CGO_ENABLED=0 go build -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-api
admin:
	CGO_ENABLED=0 go build -ldflags "-X main.octoVersion=$(VERSION) -X main.octoRevision=$(REV)" octo-admin
