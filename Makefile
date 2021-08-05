.PHONY: dev build run clean

dev: build run

build: clean
	@mkdir build && go build -o build/main .

run: build/main
	@build/main

clean:
	@rm -rf ./build