# vi:set tabstop=4 shiftwidth=4 noexpandtab:
all: install run

export GOPATH=$(PWD)

install:
	$(MAKE) build

build:
	go build counter.go storage.go main.go

run:
	go run counter.go storage.go main.go
