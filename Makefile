# vi:set tabstop=4 shiftwidth=4 noexpandtab:
all: install run

export GOPATH=$(PWD)

install:
	go get github.com/likexian/simplejson-go
	$(MAKE) build

build:
	go build counter.go storage.go main.go

run:
	go run counter.go storage.go main.go
