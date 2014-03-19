.PHONY: bin test all fmt deploy docs

all: fmt test bin

bin:
	(cd ./main; go build mcfs.go)

test:
	-go test -v ./...

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

deploy: test bin
	-cp main/mcfs $$GOPATH/bin
