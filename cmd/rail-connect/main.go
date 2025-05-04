package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sanjaykishor/rail-connect/internal/config"
	"github.com/sanjaykishor/rail-connect/internal/middleware"
	"github.com/sanjaykishor/rail-connect/internal/service"
	pb "github.com/sanjaykishor/rail-connect/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// Load configuration from config.yaml.
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger := config.NewLogger(cfg.LogLevel)

	// Create a new gRPC server.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(
			middleware.LoggingInterceptor(logger),
	)))

	sections := cfg.Sections

	// Initialize SeatManager using the configuration.
	seatManager := service.NewSeatManager(sections, logger)

	// Initialize station connection prices from config
	connectionStations := cfg.Stations

	// Initialize your service, passing the dependencies.
	ticketService := service.NewTicketManager(seatManager, connectionStations, logger)

	// Register the service with the server.
	pb.RegisterTicketBookingServiceServer(grpcServer, ticketService)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	listen, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	// Start the gRPC server in a separate goroutine.
	go func() {
		logger.Info("Server listening on", zap.String("port", cfg.Server.Port))
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logger.Info("Received signal:", zap.String("signal", sig.String()))

	logger.Info("Stopping server...")
	grpcServer.GracefulStop()
	logger.Info("Server stopped.")
}
