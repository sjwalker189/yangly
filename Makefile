run:
	go run main.go | batcat -l typescript

build:
	go build -o build/yangts
