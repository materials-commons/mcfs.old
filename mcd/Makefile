.PHONY: bin test all fmt deploy docs

all: fmt test bin

bin:
	(cd ./main; godep go build mcfs.go)

test:
	-godep go test -v ./...

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

deploy: test bin
	-cp main/mcfs $$GOPATH/bin
