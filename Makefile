.PHONY: generate clean

# Define the proto file.
PROTO_FILE = ./proto/ticketBooking.proto

# Define the output directories.
GEN_DIR = .

PROTOC = protoc

generate:
	@echo "Installing required plugins..."
	@echo "Generating code..."
	$(PROTOC) -I=. \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)
	@echo "Code generation complete!"

clean:
	@echo "Cleaning generated files..."
	find $(GEN_DIR) -name "*.pb.go" -type f -delete
	@echo "Clean complete!"
