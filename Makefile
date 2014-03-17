.PHONY: bin test all fmt deploy

all: fmt test bin

bin:
	(cd ./main; go build mcfs.go)

test:
	-go test -v ./...

fmt:
	-go fmt ./...

deploy: test bin
	-cp main/mcfs $$GOPATH/bin
