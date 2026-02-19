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

	// --------------------------------------------------------
	// Load .env file
	// --------------------------------------------------------
	// Loads environment variables from a .env file into the
	// process environment. If the file is missing, we ignore
	// the error because environment variables could be set
	// from Docker, Kubernetes, or OS-level environment.
	_ = godotenv.Load()

	// --------------------------------------------------------
	// Initialize MySQL connection using environment variables
	// --------------------------------------------------------
	// NewMySQLService() picks up DB_USER, DB_PASSWORD,
	// DB_HOST, DB_PORT, DB_NAME from the environment.
	db, err := database.NewMySQLService()
	if err != nil {
		log.Fatalf("failed to connect to mysql: %v", err)
	}

	// --------------------------------------------------------
	// Start REST server (port 8080)
	// --------------------------------------------------------
	// Runs in a separate goroutine so it doesn't block the
	// gRPC server from starting.
	go grpc.StartRESTServer(db)

	// --------------------------------------------------------
	// Start gRPC server (port 50051)
	// --------------------------------------------------------
	// This returns an object holding the server instance so
	// we can gracefully shut it down later.
	grpcServer := grpc.StartGRPCServer(db)

	// --------------------------------------------------------
	// Graceful Shutdown Handling
	// --------------------------------------------------------
	// Capture OS signals (Ctrl+C, SIGTERM) so the server can
	// shut down cleanly:
	//
	//   - Stop gRPC server gracefully
	//   - Close database connections
	//   - Close channels
	//
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	// Block until a signal is received
	<-ctx.Done()

	log.Println("Shutting down...")

	// Stop listening for further signals
	stop()

	// --------------------------------------------------------
	// Gracefully stop gRPC
	// --------------------------------------------------------
	// Allows active RPC streams to finish instead of terminating abruptly
	grpcServer.GRPC.GracefulStop()

	// --------------------------------------------------------
	// Close DB connection
	// --------------------------------------------------------
	db.Close()

	// --------------------------------------------------------
	// Close metric ingestion channel
	// --------------------------------------------------------
	close(grpc.MetricChan)

	log.Println("Shutdown complete")
}
