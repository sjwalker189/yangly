run:
	go run main.go

build:
	go build

pkg:
	go build -o npm/bin/yangly

clean:
	rm -f yangly
	rm -f npm/bin/yangly*
	rm -rf npm/node_modules
