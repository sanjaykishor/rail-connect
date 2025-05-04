.PHONY: generate clean

# Define the proto file.
PROTO_FILE = ./proto/ticketBooking.proto

# Define the output directories.
GEN_DIR = .

PROTOC = protoc

generate-proto:
	@echo "Installing required plugins..."
	@echo "Generating code..."
	$(PROTOC) -I=. \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)
	@echo "Code generation complete!"

clean-proto:
	@echo "Cleaning generated files..."
	find $(GEN_DIR) -name "*.pb.go" -type f -delete
	@echo "Clean complete!"

test:
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests complete!"

build:
	@echo "Building Go application..."
	mkdir -p ./bin
	go build -o ./bin/rail-connect ./cmd/rail-connect/main.go
	@echo "Build complete!"

run:
	@echo "Running Go application..."
	./bin/rail-connect
	@echo "Application is running!"

docker-build:
	@echo "Building Docker image..."
	docker build -t rail-connect .
	@echo "Docker image built!"

docker-run:
	@echo "Running Docker image..."
	docker run -p 50051:50051 rail-connect
	@echo "Docker image is running!"

docker-stop:
	@echo "Stopping Docker image..."
	docker stop rail-connect
	@echo "Docker image stopped!"

docker-clean:
	@echo "Cleaning Docker image..."
	docker rm rail-connect
	@echo "Docker image cleaned!"
	