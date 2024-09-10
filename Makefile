build:
	go build -o ./bin/go-wc .

run: build
	./bin/go-wc

test:
	go test ./...