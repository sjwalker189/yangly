run:
	mkdir -p out/dist/types # TODO
	go run main.go
build:
	go build -o build/yangts

clean:
	rm -r out/
