run:
	mkdir -p out/dist/types # TODO
	go run main.go

build:
	go build

clean:
	rm -r out/
