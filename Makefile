# Makefile

# Variables
BINARY_NAME=proxy-server
DOCKER_IMAGE_NAME=go-caching-proxy

# Phony targets don't represent actual files
.PHONY: run build test docker-build docker-run clean

# Runs the server using the local config
run:
	go run ./cmd/server

# Builds the binary for the current OS/ARCH into the ./bin directory
build:
	go build -o ./bin/${BINARY_NAME} ./cmd/server

# Run tests for all packages
test:
	go test -v ./...

# Build the production Docker image
docker-build:
	docker build -t ${DOCKER_IMAGE_NAME}:latest .

# Run the application inside a Docker container
docker-run:
	docker run -p 8080:8080 --rm --name ${DOCKER_IMAGE_NAME} ${DOCKER_IMAGE_NAME}:latest

# Removes the build artifacts
clean:
	rm -rf ./bin