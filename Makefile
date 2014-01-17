.PHONY: bin test all fmt

all: fmt test bin

bin:
	(cd main; go build mcfs.go)

test:
	-go test ./...

fmt:
	-go fmt ./...
