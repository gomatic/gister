build: test
	go build -o bin/gist ./cmd/...

test:
	go test ./...