build:
	CGO_ENABLED=0 go build -o ./bin/blockchain .

run: build
	./bin/blockchain

test:
	go test -v ./...
