# vi:set tabstop=4 shiftwidth=4 noexpandtab:
all: install run

install:
	$(MAKE) build

build:
	go build -o counter

run:
	go run main.go

clean:
	rm -f counter
	rm -f data/counter.json
