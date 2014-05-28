.PHONY: all test fmt docs

all: fmt test docs

test:
	rm -rf test_data/t
	-godep go test -v ./...

docs:
	./makedocs.sh

fmt:
	-go fmt ./...
