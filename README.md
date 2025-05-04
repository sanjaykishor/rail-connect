# Train Ticket Booking Service - gRPC API

## Overview
The **Train Ticket Booking Service** is a gRPC-based system that allows users to book tickets, retrieve receipts, manage seat assignments, and cancel bookings for rail travel between London and France.

The system implements seat management with an optimized round-robin allocation strategy across multiple train sections, ensuring balanced seat distribution and efficient resource utilization.

## gRPC Service Definition
The service is defined in the **ticketBooking.proto** file and includes the following RPC methods:

### **TicketBookingService**
```proto
service TicketBookingService {
  rpc PurchaseTicket(PurchaseTicketRequest) returns (PurchaseTicketResponse) {};
  rpc GetReceipt(GetReceiptRequest) returns (GetReceiptResponse) {};
  rpc GetUsersBySection(GetUsersBySectionRequest) returns (GetUsersBySectionResponse) {};
  rpc RemoveUser(RemoveUserRequest) returns (RemoveUserResponse) {};
  rpc UpdateUserSeat(UpdateUserSeatRequest) returns (UpdateUserSeatResponse) {};
}
```

## Features
### **1. Ticket Management**
- **PurchaseTicket:** Allows users to purchase tickets and assigns them a seat
- **GetReceipt:** Retrieves the ticket receipt for a specific user
- **GetUsersBySection:** Retrieves all users seated in a specific section
- **RemoveUser:** Cancels a ticket and releases the assigned seat
- **UpdateUserSeat:** Allows users to change their seat allocation

### **2. Seat Management**
- **Seat allocation:** Seats are assigned in a round-robin manner across sections
- **Seat modification:** Users can request to change their assigned seats
- **Seat release:** When a ticket is canceled, the seat becomes available again

## Messages Definition

### **User Information**
```proto
message User {
  string firstName = 1;
  string lastName = 2;
  string email = 3;
}
```

### **Ticket Booking Requests & Responses**
```proto
message PurchaseTicketRequest {
  User user = 1;
  string from = 4;
  string to = 5;
}

message PurchaseTicketResponse {
  string message = 1;
  Receipt receipt = 2;
}

message Receipt {
  string from = 1;
  string to = 2;
  User user = 3;
  double pricePaid = 4;
  Seat seat = 5;
}
```

### **Seat Management**
```proto
message Seat {
  string section = 1;
  int32 seatNumber = 2;
}
```

### **Ticket Lookup & Cancellation**
```proto
message GetReceiptRequest {
  string email = 1;
}

message GetReceiptResponse {
  Receipt receipt = 1;
}

message RemoveUserRequest {
  string email = 1;
}

message RemoveUserResponse {
  string message = 1;
  User removedUser = 2;
}
```

### **Section-wise User Retrieval**
```proto
message GetUsersBySectionRequest {
  string section = 1;
}

message UserSeat {
  User user = 1;
  int32 allottedSeat = 2;
}

message GetUsersBySectionResponse {
  string section = 1;
  repeated UserSeat users = 2;
}
```

### **Seat Modification**
```proto
message UpdateUserSeatRequest {
  string email = 1;
  Seat newSeat = 2;
}

message UpdateUserSeatResponse {
  string message = 1;
  Receipt updatedReceipt = 2;
}
```

## Repository Structure

```
rail-connect/
├── cmd/                    # Application entry points
│   └── rail-connect/       # Main server application
├── internal/               # Internal packages
│   ├── config/             # Configuration handling
│   ├── middleware/         # gRPC middleware
│   └── service/            # Core business logic
├── proto/                  # Protocol Buffer definitions
├── client/                 # Example client implementation
├── config/                 # Configuration files
├── Dockerfile              # Docker build definition
├── Makefile                # Build automation
└── README.md               # Documentation
```

## Architecture

Rail-Connect is built using Go and follows a clean, modular architecture:

- **gRPC Service Layer**: Handles client requests and responses
- **Ticket Manager**: Core business logic for ticket operations
- **Seat Manager**: Optimized seat allocation using round-robin across sections
- **Configuration**: YAML-based configuration for sections, pricing, and server settings
- **Middleware**: Request logging and interceptors

## Running the Service

### **1. Install Dependencies**

Ensure you have Go installed (version 1.18 or later) and the required gRPC tools:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### **2. Clone the Repository**

```sh
git clone https://github.com/sanjaykishor/rail-connect.git
cd rail-connect
```

### **3. Install Go Dependencies**

```sh
go mod download
```

### **4. Generate Protocol Buffer Code**

```sh
make generate-proto
```

### **5. Build the Application**

```sh
make build
```

### **6. Run the Server**

```sh
make run
# or directly with:
./bin/rail-connect
```

### **7. Docker Deployment**

#### Building the Docker Image:

```sh
# Using the Makefile
make docker-build

# Or directly with Docker
docker build -t rail-connect .
```

#### Running with Docker:

```sh
# Basic usage
docker run -p 50051:50051 rail-connect

# Run in detached mode
docker run -d -p 50051:50051 rail-connect

# With a custom name
docker run -d -p 50051:50051 --name rail-connect-service rail-connect
```

### **8. Running the Example Client**

A complete example client implementation is provided:

```sh
go run client/example.go
```

### **9. Running Tests**

```sh
make test
```

## API Usage Examples

### **1. Purchase a Ticket**

```bash
grpcurl -plaintext -d '{
  "user": {
    "firstName": "Sanjay",
    "lastName": "Kishor",
    "email": "test@example.com"
  },
  "from": "London",
  "to": "France"
}' localhost:50051 ticketBooking.TicketBookingService/PurchaseTicket
```

### **2. Get Receipt**

```bash
grpcurl -plaintext -d '{
  "email": "test@example.com"
}' localhost:50051 ticketBooking.TicketBookingService/GetReceipt
```

### **3. View Users by Section**

```bash
grpcurl -plaintext -d '{
  "section": "A"
}' localhost:50051 ticketBooking.TicketBookingService/GetUsersBySection
```

### **4. Update User Seat**

```bash
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "newSeat": {
    "section": "B",
    "seatNumber": 25
  }
}' localhost:50051 ticketBooking.TicketBookingService/UpdateUserSeat
```

### **5. Remove User**

```bash
grpcurl -plaintext -d '{
  "email": "test@example.com"
}' localhost:50051 ticketBooking.TicketBookingService/RemoveUser
```


