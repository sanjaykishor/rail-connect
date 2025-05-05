package main

import (
	"context"
	"flag"
	"log"

	"github.com/sanjaykishor/rail-connect/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	address = flag.String("address", "localhost:50051", "The server address in the format of host:port")
)

func main() {
	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}
	defer conn.Close()

	client := proto.NewTicketBookingServiceClient(conn)

	// Purchase a ticket
	user1 := &proto.User{
		Email:     "test1@example.com",
		FirstName: "Sanjay",
		LastName:  "Kishor",
	}

	purchaseRes1, err := client.PurchaseTicket(context.Background(), &proto.PurchaseTicketRequest{
		User: user1,
		From: "London",
		To:   "France",
	})
	if err != nil {
		log.Fatalf("could not purchase ticket: %v\n", err)
	}
	log.Printf("Ticket purchased successfully: %v\n", purchaseRes1.Receipt)

	user2 := &proto.User{
		Email:     "test2@example.com",
		FirstName: "Sanjay",
		LastName:  "Kishor",
	}

	purchaseRes2, err := client.PurchaseTicket(context.Background(), &proto.PurchaseTicketRequest{
		User: user2,
		From: "London",
		To:   "France",
	})
	if err != nil {
		log.Fatalf("could not purchase ticket: %v\n", err)
	}
	log.Printf("Ticket purchased successfully: %v\n", purchaseRes2.Receipt)

	user3 := &proto.User{
		Email:     "test3@example.com",
		FirstName: "Sanjay",
		LastName:  "Kishor",
	}

	purchaseRes3, err := client.PurchaseTicket(context.Background(), &proto.PurchaseTicketRequest{
		User: user3,
		From: "London",
		To:   "France",
	})

	if err != nil {
		log.Fatalf("could not purchase ticket: %v\n", err)
	}

	log.Printf("Ticket purchased successfully: %v\n", purchaseRes3.Receipt)

	// Get the ticket for a user
	getTicketRes, err := client.GetReceipt((context.Background()), &proto.GetReceiptRequest{
		Email: user3.Email,
	})
	if err != nil {
		log.Fatalf("could not get ticket: %v\n", err)
	}
	log.Printf("Ticket retrieved successfully: %v\n", getTicketRes.Receipt)

	// get Users by section
	getUsersRes, err := client.GetUsersBySection(context.Background(), &proto.GetUsersBySectionRequest{
		Section: "A",
	})

	if err != nil {
		log.Fatalf("could not get users by section: %v\n", err)
	}

	log.Printf("Users in section %s: %v\n", getUsersRes.Section, getUsersRes.Users)

	// Update the user's seat
	updateRes, err := client.UpdateUserSeat((context.Background()), &proto.UpdateUserSeatRequest{
		Email: user1.Email,
		NewSeat: &proto.Seat{
			Section:    getTicketRes.Receipt.Seat.Section,
			SeatNumber: getTicketRes.Receipt.Seat.SeatNumber + 1,
		},
	})

	if err != nil {
		log.Fatalf("could not update user seat: %v\n", err)
	}
	log.Printf("User seat updated successfully: %v\n", updateRes)

	// Remove the user's
	removeRes, err := client.RemoveUser((context.Background()), &proto.RemoveUserRequest{
		Email: user1.Email,
	})
	if err != nil {
		log.Fatalf("could not remove user: %v\n", err)
	}
	log.Printf("User removed successfully: %v\n", removeRes)
}
