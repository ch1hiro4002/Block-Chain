build:
	go build -o ./bin/Block-Chain

run: build
	./bin/Block-Chain

test:
	go test -v ./...