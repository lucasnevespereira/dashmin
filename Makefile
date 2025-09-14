.PHONY: build install clean run setup test

# Build the binary
build:
	go build -o bin/dashmin main.go

# Install to GOPATH/bin
install:
	go install .

# Clean build artifacts
clean:
	rm -rf bin/

# Run the dashboard
run:
	go run main.go

# Run setup
setup:
	go run main.go setup

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/dashmin-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/dashmin-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/dashmin-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/dashmin-windows-amd64.exe main.go
