package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gowatch/internal/database"
	"gowatch/internal/grpc"
)

func main() {

	// Connect to MySQL
	db, err := database.NewMySQLService("root", "rootroot", "127.0.0.1:3306", "gowatch")
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

	stop()

	// Graceful gRPC
	grpcServer.GRPC.GracefulStop()

	// Close DB
	db.Close()

	// Close channel
	close(grpc.MetricChan)

	log.Println("Shutdown complete")
}
