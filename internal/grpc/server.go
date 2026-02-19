package grpc

import (
	pb "gowatch/gopherwatch/pkg/generated" // Generated protobuf & gRPC bindings
	"gowatch/internal/database"            // Database service layer
	"io"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

// Global DB reference (optional depending on design)
// Currently unused but can store DB instance for server-wide access if needed.
var globalDB database.Service

// -------------------- gRPC Metrics Server --------------------

// MetricsServer implements the gRPC MetricsService defined in protobufs.
type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer // Required for forward compatibility
}

// SendMetrics handles a bidirectional streaming RPC where clients
// continuously send metric data to the server.
//
// This function receives metrics in a loop, pushes them into MetricChan
// (for worker processing), and logs each message.
func (s *MetricsServer) SendMetrics(stream pb.MetricsService_SendMetricsServer) error {

	for {
		// Receive a metric message from the stream
		metric, err := stream.Recv()

		// If the client closes the stream normally, return a summary
		if err == io.EOF {
			return stream.SendAndClose(&pb.Summary{
				Message: "Received metrics successfully",
			})
		}

		// If the stream errors out unexpectedly
		if err != nil {
			return err
		}

		// FAN-OUT PATTERN:
		// Send received metric into a buffered global channel
		// Workers consume and evaluate alert rules.
		MetricChan <- metric

		// Log received CPU value
		log.Printf("Received from Agent: %s | CPU: %.2f",
			metric.AgentId,
			metric.CpuUsage,
		)
	}
}

// -------------------- gRPC Server Bootstrap --------------------

type ServerInstance struct {
	GRPC *grpc.Server
}

// StartGRPCServer starts the gRPC server on port 50051,
// initializes workers, sets up interceptors, and registers RPC services.
func StartGRPCServer(db database.Service) *ServerInstance {
	// Start N worker goroutines for processing metrics
	StartWorkers(10, db)

	// Listen on TCP port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	// Create gRPC server with optional interceptors
	server := grpc.NewServer(
		grpc.StreamInterceptor(ServiceIDInterceptor), // Adds metadata to logs
	)

	// Register the MetricsService implementation
	pb.RegisterMetricsServiceServer(server, &MetricsServer{})

	// Launch gRPC server in a separate goroutine
	go func() {
		log.Println("gRPC Server running on :50051")
		if err := server.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// Return reference to server for graceful shutdown
	return &ServerInstance{
		GRPC: server,
	}
}

// -------------------- REST Server --------------------

type RestServer struct {
	db database.Service
}

// StartRESTServer launches the REST HTTP server on port 8080.
// It registers all HTTP routes through the router defined in routes.go.
func StartRESTServer(db database.Service) {
	srv := &RestServer{db: db}

	// Start HTTP server with registered routes
	err := http.ListenAndServe(":8080", srv.RegisterRoutes())
	if err != nil {
		log.Fatalf("REST server error: %v", err)
	}
}
