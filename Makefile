# vi:set tabstop=4 shiftwidth=4 noexpandtab:
all: deps build run

deps:
	test -d vendor/github.com/op/go-logging || git clone --depth 1 git@github.com:op/go-logging vendor/github.com/op/go-logging

build:
	go build -o bin/counter

run:
	go run main.go

clean:
	rm -f bin/counter
	rm -f data/counter.json
