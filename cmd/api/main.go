package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gowatch/internal/database"
	"gowatch/internal/grpc"

	"github.com/joho/godotenv"
)

func main() {

	// Load .env file (optional but recommended)
	_ = godotenv.Load()

	// Connect to MySQL using env variables
	db, err := database.NewMySQLService()
	if err != nil {
		log.Fatalf("failed to connect to mysql: %v", err)
	}

	// Start REST server
	go grpc.StartRESTServer(db)

	// Start gRPC server
	grpcServer := grpc.StartGRPCServer(db)

	// Handle shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()

	log.Println("Shutting down...")

	// Stop receiving signals
	stop()

	// Graceful gRPC shutdown
	grpcServer.GRPC.GracefulStop()

	// Close DB connection
	db.Close()

	// Close the metric channel
	close(grpc.MetricChan)

	log.Println("Shutdown complete")
}
