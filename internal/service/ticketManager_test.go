package service

import (
	"context"
	"testing"

	"github.com/sanjaykishor/rail-connect/internal/config"
	"github.com/stretchr/testify/assert"

	pb "github.com/sanjaykishor/rail-connect/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
)

func createTestTicketManager() *TicketManager {
	sections := []config.SectionConfig{
		{Name: "A", MaxSeats: 20},
		{Name: "B", MaxSeats: 20},
	}
	logger, _ := zap.NewProduction()
	seatManager := NewSeatManager(sections, logger)
	connectionStations := map[string]float64{
		"London-France": 20.00,
	}
	return NewTicketManager(seatManager, connectionStations, logger)
}

func TestNewTicketManager(t *testing.T) {
	tm := createTestTicketManager()
	assert.NotNil(t, tm, "Expected TicketManager to be created")
	assert.NotNil(t, tm.SeatManager, "Expected SeatManager to be initialized")
	assert.NotNil(t, tm.StationConnection, "Expected StationConnection to be initialized")
	assert.NotNil(t, tm.Receipts, "Expected Receipts map to be initialized")
}

func TestBookTicket(t *testing.T) {
	tm := createTestTicketManager()

	tests := []struct {
		name          string
		request       *pb.PurchaseTicketRequest
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "Valid Request",
			request: &pb.PurchaseTicketRequest{
				User: &pb.User{
					Email:     "test1@example.com",
					FirstName: "Sanjay",
					LastName:  "Kishor",
				},
				From: "London",
				To:   "France",
			},
			expectedError: false,
			expectedCode:  codes.OK,
		},
		{
			name: "Invalid Request - Missing User",
			request: &pb.PurchaseTicketRequest{
				From: "London",
				To:   "France",
			},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "Invalid Request - Missing From",
			request: &pb.PurchaseTicketRequest{
				User: &pb.User{
					Email:     "test2@example.com",
					FirstName: "Sanjay",
					LastName:  "Kishor",
				},
				To: "France",
			},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "Invalid Request - Missing To",
			request: &pb.PurchaseTicketRequest{
				User: &pb.User{
					Email:     "test3@example.com",
					FirstName: "Sanjay",
					LastName:  "Kishor",
				},
				From: "London",
			},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "Invalid Request - Invalid Station",
			request: &pb.PurchaseTicketRequest{
				User: &pb.User{
					Email:     "test4@example.com",
					FirstName: "Sanjay",
					LastName:  "Kishor",
				},
			},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := tm.BookTicket(context.Background(), test.request)
			if test.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, st.Code())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.Receipt)
				assert.Equal(t, response.Message, "Ticket booked successfully")
			}
		})
	}

}

func TestGetReceipt(t *testing.T) {
	tm := createTestTicketManager()

	userEmail := "test@example.com"
	tm.Receipts[userEmail] = &pb.Receipt{
		User:      &pb.User{FirstName: "Sanjay", LastName: "Kishor", Email: userEmail},
		Seat:      &pb.Seat{Section: "A", SeatNumber: 1},
		From:      "London",
		To:        "France",
		PricePaid: 20.00,
	}

	tests := []struct {
		name          string
		request       *pb.GetReceiptRequest
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "Valid Request",
			request: &pb.GetReceiptRequest{
				Email: userEmail,
			},
			expectedError: false,
			expectedCode:  codes.OK,
		},
		{
			name:          "Invalid Request - Missing Email",
			request:       &pb.GetReceiptRequest{},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "Invalid Request - Nonexistent Email",
			request: &pb.GetReceiptRequest{
				Email: "nonexist@example.com",
			},
			expectedError: true,
			expectedCode:  codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := tm.GetReceipt(context.Background(), test.request)
			if test.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, st.Code())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.Receipt)
				assert.Equal(t, response.Receipt.User.Email, userEmail)
			}
		})
	}
}

func TestGetUsersBySection(t *testing.T) {
	tm := createTestTicketManager()

	tm.Receipts["test1@example.com"] = &pb.Receipt{
		User:      &pb.User{FirstName: "Sanjay", LastName: "Kishor", Email: "test1@example.com"},
		Seat:      &pb.Seat{Section: "A", SeatNumber: 1},
		From:      "London",
		To:        "France",
		PricePaid: 20.00,
	}

	tm.Receipts["test2@example.com"] = &pb.Receipt{
		User:      &pb.User{FirstName: "Sanjay", LastName: "Kishor", Email: "test2@example.com"},
		Seat:      &pb.Seat{Section: "A", SeatNumber: 2},
		From:      "London",
		To:        "France",
		PricePaid: 20.00,
	}

	tests := []struct {
		name          string
		request       *pb.GetUsersBySectionRequest
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name:          "Valid Request - Section A",
			request:       &pb.GetUsersBySectionRequest{Section: "A"},
			expectedError: false,
			expectedCode:  codes.OK,
		},
		{
			name:          "Invalid Request - Missing Section",
			request:       &pb.GetUsersBySectionRequest{},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name:          "Invalid Request - Nonexistent Section",
			request:       &pb.GetUsersBySectionRequest{Section: "C"},
			expectedError: true,
			expectedCode:  codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := tm.GetUsersBySection(context.Background(), test.request)
			if test.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, st.Code())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.Users)
				assert.Equal(t, response.Section, "A")
			}
		})
	}
}

func TestUpdateUserSeat(t *testing.T) {
	tm := createTestTicketManager()

	userEmail := "test@example.com"
	seatNumber, section := 1, "A"

	// assign the seat
	tm.SeatManager.Sections[section].Seats[seatNumber] = &Seat{
		Number:    seatNumber,
		Available: false,
	}

	tm.Receipts[userEmail] = &pb.Receipt{
		User:      &pb.User{FirstName: "Sanjay", LastName: "Kishor", Email: userEmail},
		Seat:      &pb.Seat{Section: section, SeatNumber: int32(seatNumber)},
		From:      "London",
		To:        "France",
		PricePaid: 20.00,
	}

	tests := []struct {
		name          string
		request       *pb.UpdateUserSeatRequest
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "Valid Request",
			request: &pb.UpdateUserSeatRequest{
				Email: userEmail,
				NewSeat: &pb.Seat{
					Section:    section,
					SeatNumber: int32(seatNumber + 1),
				},
			},
			expectedError: false,
			expectedCode:  codes.OK,
		},
		{
			name: "Invalid Request - Missing Email",
			request: &pb.UpdateUserSeatRequest{
				NewSeat: &pb.Seat{
					Section:    section,
					SeatNumber: int32(seatNumber + 1),
				},
			},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "Invalid Request - Missing NewSeat",
			request: &pb.UpdateUserSeatRequest{
				Email: userEmail,
			},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "Invalid Request - Nonexistent Email",
			request: &pb.UpdateUserSeatRequest{
				Email: userEmail + "nonexist",
				NewSeat: &pb.Seat{
					Section:    section,
					SeatNumber: int32(seatNumber + 1),
				},
			},
			expectedError: true,
			expectedCode:  codes.NotFound,
		},
		{
			name: "Invalid Request - Nonexistent Section",
			request: &pb.UpdateUserSeatRequest{
				Email: userEmail,
				NewSeat: &pb.Seat{
					Section:    "C",
					SeatNumber: int32(seatNumber + 1),
				},
			},
			expectedError: true,
			expectedCode:  codes.NotFound,
		},
		{
			name: "Invalid Request - Nonexistent Seat",
			request: &pb.UpdateUserSeatRequest{
				Email: userEmail,
				NewSeat: &pb.Seat{
					Section:    section,
					SeatNumber: int32(seatNumber + 100),
				},
			},
			expectedError: true,
			expectedCode:  codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := tm.UpdateSeat(context.Background(), test.request)
			if test.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, st.Code())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, response.Message, "Seat updated successfully")
			}
		})
	}

}

func TestRemoveUser(t *testing.T) {
	tm := createTestTicketManager()

	userEmail := "test@example.com"
	seatNumber, section := 1, "A"

	// assign the seat
	tm.SeatManager.Sections[section].Seats[seatNumber] = &Seat{
		Number:    seatNumber,
		Available: false,
	}

	tm.Receipts[userEmail] = &pb.Receipt{
		User:      &pb.User{FirstName: "Sanjay", LastName: "Kishor", Email: userEmail},
		Seat:      &pb.Seat{Section: section, SeatNumber: int32(seatNumber)},
		From:      "London",
		To:        "France",
		PricePaid: 20.00,
	}

	tests := []struct {
		name          string
		request       *pb.RemoveUserRequest
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "Valid Request",
			request: &pb.RemoveUserRequest{
				Email: userEmail,
			},
			expectedError: false,
			expectedCode:  codes.OK,
		},
		{
			name:          "Invalid Request - Missing Email",
			request:       &pb.RemoveUserRequest{},
			expectedError: true,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "Invalid Request - Nonexistent Email",
			request: &pb.RemoveUserRequest{
				Email: userEmail + "nonexist",
			},
			expectedError: true,
			expectedCode:  codes.NotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := tm.RemoveUser(context.Background(), test.request)
			if test.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, st.Code())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, response.Message, "Ticket cancelled successfully")
			}
		})
	}
}
