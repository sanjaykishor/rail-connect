package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	pb "github.com/sanjaykishor/rail-connect/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TicketManager handles ticket purchases, retrievals, and modifications.
// It interacts with SeatManager to manage seat assignments for tickets.
type TicketManager struct {
	pb.UnimplementedTicketBookingServiceServer
	SeatManager       *SeatManager
	Receipts          map[string]*pb.Receipt
	mu                sync.Mutex
	StationConnection map[string]float64
	Logger            *zap.Logger
}

// NewTicketManager creates a new TicketManager with the given seat manager and connection stations
// and initializes the receipts map.
func NewTicketManager(seatManager *SeatManager, connectionStations map[string]float64, logger *zap.Logger) *TicketManager {
	return &TicketManager{
		SeatManager:       seatManager,
		StationConnection: connectionStations,
		Receipts:          make(map[string]*pb.Receipt),
		Logger:            logger,
	}
}

// PurchaseTicket processes a ticket purchase request, assigns a seat, and returns a ticket receipt.
func (tm *TicketManager) PurchaseTicket(ctx context.Context, req *pb.PurchaseTicketRequest) (*pb.PurchaseTicketResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.Logger.Info("PurchaseTicket request received")

	// Validate the request
	if req == nil {
		tm.Logger.Error("PurchaseTicket request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	// Check if the user is valid
	if req.User == nil || req.User.Email == "" || req.From == "" || req.To == "" {
		fields := []zap.Field{
			zap.String("from", req.From),
			zap.String("to", req.To),
		}

		if req.User != nil {
			fields = append(fields, zap.String("user_email", req.User.Email))
		} else {
			fields = append(fields, zap.String("user", "<nil>"))
		}

		tm.Logger.Error("PurchaseTicket request missing required fields", fields...)
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	// TODO: To be decided if we want to allow multiple tickets for the same user
	// if _, exists := tm.Receipts[req.User.Email]; exists {
	// 	tm.Logger.Error("User already has a ticket",
	// 		zap.String("user", req.User.Email),
	// 	)
	// 	return nil, status.Error(codes.AlreadyExists, "User already has a ticket")
	// }

	tm.Logger.Info("PurchaseTicket request",
		zap.String("user", req.User.Email),
		zap.String("from", req.From),
		zap.String("to", req.To),
		zap.Time("timestamp", time.Now()),
	)

	// Validate the station names
	connectionStations := fmt.Sprintf("%s-%s", req.From, req.To)
	if tm.StationConnection[connectionStations] == 0 {
		tm.Logger.Error("PurchaseTicket invalid station names",
			zap.String("from", req.From),
			zap.String("to", req.To),
			zap.String("connection", connectionStations),
		)
		return nil, status.Error(codes.InvalidArgument, "invalid station")
	}

	section, seat, err := tm.SeatManager.AssignSeat()
	if err != nil {
		tm.Logger.Error("PurchaseTicket failed to assign seat",
			zap.String("user", req.User.Email),
			zap.String("from", req.From),
			zap.String("to", req.To),
			zap.Error(err),
		)
		return nil, status.Error(codes.NotFound, "failed to assign seat")
	}

	receipt := &pb.Receipt{
		User:      req.User,
		From:      req.From,
		To:        req.To,
		PricePaid: tm.StationConnection[connectionStations],
		Seat:      &pb.Seat{SeatNumber: int32(seat), Section: section},
	}

	tm.Receipts[req.User.Email] = receipt

	tm.Logger.Info("PurchaseTicket successful",
		zap.String("user", req.User.Email),
		zap.String("from", req.From),
		zap.String("to", req.To),
		zap.Int("seat_number", seat),
		zap.String("section", section),
		zap.Float64("price_paid", tm.StationConnection[connectionStations]),
	)
	return &pb.PurchaseTicketResponse{
		Message: "Ticket booked successfully",
		Receipt: receipt,
	}, nil

}

// GetReceipt retrieves the ticket receipt for a user based on their email
func (tm *TicketManager) GetReceipt(ctx context.Context, req *pb.GetReceiptRequest) (*pb.GetReceiptResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Logger.Info("GetReceipt request received")

	// Validate the request
	if req == nil {
		tm.Logger.Error("GetReceipt request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	// Check if the user is valid
	if req.Email == "" {
		tm.Logger.Error("GetReceipt request missing required fields",
			zap.String("email", req.Email),
		)
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	tm.Logger.Info("GetReceipt request",
		zap.String("email", req.Email),
		zap.Time("timestamp", time.Now()),
	)

	receipt, exists := tm.Receipts[req.Email]
	if !exists {
		tm.Logger.Error("GetReceipt ticket receipt not found",
			zap.String("email", req.Email),
		)
		return nil, status.Error(codes.NotFound, "ticket receipt not found")
	}

	tm.Logger.Info("GetReceipt successful",
		zap.String("email", req.Email),
		zap.String("from", receipt.From),
		zap.String("to", receipt.To),
		zap.Int("seat_number", int(receipt.Seat.SeatNumber)),
		zap.String("section", receipt.Seat.Section),
		zap.Float64("price_paid", receipt.PricePaid),
	)
	return &pb.GetReceiptResponse{
		Receipt: receipt,
	}, nil
}

// GetUsersBySection retrieves all users in a specific section and their seats
func (tm *TicketManager) GetUsersBySection(ctx context.Context, req *pb.GetUsersBySectionRequest) (*pb.GetUsersBySectionResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Logger.Info("GetUsersBySection request received")

	// Validate the request
	if req == nil {
		tm.Logger.Error("GetUsersBySection request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	// Check if the user is valid
	if req.Section == "" {
		tm.Logger.Error("GetUsersBySection request missing required fields",
			zap.String("section", req.Section),
		)
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	// Check if the section exists
	if _, exists := tm.SeatManager.Sections[req.Section]; !exists {
		tm.Logger.Error("GetUsersBySection section not found",
			zap.String("section", req.Section),
		)
		return nil, status.Error(codes.NotFound, "section not found")
	}

	tm.Logger.Info("GetUsersBySection request",
		zap.String("section", req.Section),
		zap.Time("timestamp", time.Now()),
	)

	users := make([]*pb.UserSeat, 0)
	for _, receipt := range tm.Receipts {
		if receipt.Seat.Section == req.Section {
			users = append(users, &pb.UserSeat{
				User:         receipt.User,
				AllottedSeat: receipt.Seat.SeatNumber,
			})
		}
	}

	tm.Logger.Info("GetUsersBySection successful",
		zap.String("section", req.Section),
		zap.Int("user_count", len(users)),
	)

	return &pb.GetUsersBySectionResponse{
		Section: req.Section,
		Users:   users,
	}, nil
}

// UpdateUserSeat changes the seat assignment for a user.
func (tm *TicketManager) UpdateUserSeat(ctx context.Context, req *pb.UpdateUserSeatRequest) (*pb.UpdateUserSeatResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Logger.Info("UpdateUserSeat request received")

	// Validate the request
	if req == nil {
		tm.Logger.Error("UpdateUserSeat request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	// Check if the user is valid
	if req.Email == "" || req.NewSeat == nil || req.NewSeat.Section == "" || req.NewSeat.SeatNumber == 0 {
		fields := []zap.Field{
			zap.String("email", req.Email),
		}

		if req.NewSeat != nil {
			fields = append(fields, zap.String("new_section", req.NewSeat.Section))
			fields = append(fields, zap.Int32("new_seat", req.NewSeat.SeatNumber))
		} else {
			fields = append(fields, zap.String("new_section", "<nil>"))
			fields = append(fields, zap.String("new_seat", "<nil>"))
		}
		tm.Logger.Error("UpdateUserSeat request missing required fields", fields...)

		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	tm.Logger.Info("UpdateUserSeat request",
		zap.String("email", req.Email),
		zap.String("new_section", req.NewSeat.Section),
		zap.Int32("new_seat", req.NewSeat.SeatNumber),
		zap.Time("timestamp", time.Now()),
	)

	receipt, exists := tm.Receipts[req.Email]
	if !exists {
		tm.Logger.Error("UpdateUserSeat ticket receipt not found",
			zap.String("email", req.Email),
		)
		return nil, status.Error(codes.NotFound, "ticket receipt not found")
	}

	if err := tm.SeatManager.UpdateSeat(int(receipt.Seat.SeatNumber), receipt.Seat.Section, int(req.NewSeat.SeatNumber), req.NewSeat.Section); err != nil {
		tm.Logger.Error("UpdateUserSeat failed to update seat",
			zap.String("email", req.Email),
			zap.String("new_section", req.NewSeat.Section),
			zap.Int32("new_seat", req.NewSeat.SeatNumber),
			zap.Error(err),
		)
		return nil, status.Error(codes.NotFound, "failed to update seat")
	}

	receipt.Seat = req.NewSeat

	tm.Logger.Info("UpdateUserSeat successful",
		zap.String("email", req.Email),
		zap.String("new_section", req.NewSeat.Section),
		zap.Int32("new_seat", req.NewSeat.SeatNumber),
		zap.Float64("price_paid", receipt.PricePaid),
	)
	return &pb.UpdateUserSeatResponse{
		Message:        "Seat updated successfully",
		UpdatedReceipt: receipt,
	}, nil
}

// RemoveUser cancels a user's ticket and releases the seat
func (tm *TicketManager) RemoveUser(ctx context.Context, req *pb.RemoveUserRequest) (*pb.RemoveUserResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Logger.Info("RemoveUser request received")

	// Validate the request
	if req == nil {
		tm.Logger.Error("RemoveUser request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	// Check if the user is valid
	if req.Email == "" {
		tm.Logger.Error("RemoveUser request missing required fields",
			zap.String("email", req.Email),
		)
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	tm.Logger.Info("RemoveUser request",
		zap.String("email", req.Email),
		zap.Time("timestamp", time.Now()),
	)

	receipt, exists := tm.Receipts[req.Email]
	if !exists {
		tm.Logger.Error("RemoveUser ticket receipt not found",
			zap.String("email", req.Email),
		)
		return nil, status.Error(codes.NotFound, "ticket receipt not found")
	}

	// Store user before removing
	user := receipt.User

	if err := tm.SeatManager.ReleaseSeat(receipt.Seat.Section, int(receipt.Seat.SeatNumber)); err != nil {
		tm.Logger.Error("RemoveUser failed to release seat",
			zap.String("email", req.Email),
			zap.String("section", receipt.Seat.Section),
			zap.Int32("seat_number", receipt.Seat.SeatNumber),
			zap.Error(err),
		)
		return nil, status.Error(codes.NotFound, "failed to release seat")
	}

	delete(tm.Receipts, req.Email)

	tm.Logger.Info("RemoveUser successful",
		zap.String("email", req.Email),
		zap.String("section", receipt.Seat.Section),
		zap.Int32("seat_number", receipt.Seat.SeatNumber),
	)
	return &pb.RemoveUserResponse{
		Message:     "Ticket cancelled successfully",
		RemovedUser: user,
	}, nil
}
