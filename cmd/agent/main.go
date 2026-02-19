package main

import (
	"context"
	"log"
	"time"

	pb "gowatch/gopherwatch/pkg/generated" // Import generated gRPC protobuf code

	"github.com/shirou/gopsutil/v3/cpu" // Library for retrieving system CPU metrics
	"google.golang.org/grpc"            // gRPC client package
)

func main() {
	// Establish a gRPC connection to the server.
	// `WithInsecure` is used because there is no TLS configured.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close() // Ensure connection closes on exit

	// Initialize the gRPC client using the generated protobuf interface.
	client := pb.NewMetricsServiceClient(conn)

	// Open a streaming RPC to send metrics continuously.
	stream, err := client.SendMetrics(context.Background())
	if err != nil {
		log.Fatalf("failed to open stream: %v", err)
	}

	// Infinite loop to periodically collect and send metrics.
	for {
		// Retrieve CPU usage percentage.
		// The first argument is the sampling duration (1 second).
		// The second argument `false` means we want aggregated CPU usage across all cores.
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			log.Printf("cpu error: %v", err)
			continue // Continue instead of failing, to keep the agent alive
		}

		// Create a metric report message to send via gRPC streaming.
		metric := &pb.MetricReport{
			AgentId:     "macbook-pro",     // Identifier for this agent
			Timestamp:   time.Now().Unix(), // Unix timestamp for when the reading is taken
			CpuUsage:    percent[0],        // CPU usage value (first element from array)
			MemoryUsage: 0,                 // Memory & disk usage placeholders for now
			DiskUsage:   0,
		}

		// Log what we're sending for visibility.
		log.Printf("Sending CPU = %.2f%%", percent[0])

		// Send the metric to the server through the open stream.
		if err := stream.Send(metric); err != nil {
			log.Fatalf("send error: %v", err)
		}
	}
}
