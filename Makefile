

build:
	make deps
	GOPATH=$(shell readlink -f .) go build -o docker-hipache-updater

deps:
	GOPATH=$(shell readlink -f .) go get -d ./...

clean:
	rm -rf src


run:
	GOPATH=$(shell readlink -f .) go run docker-hipache-updater.go --config config.sample.json

