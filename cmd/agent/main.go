package main

import (
	"context"
	"log"
	"time"

	pb "gowatch/gopherwatch/pkg/generated"

	"github.com/shirou/gopsutil/v3/cpu"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMetricsServiceClient(conn)

	stream, err := client.SendMetrics(context.Background())
	if err != nil {
		log.Fatalf("failed to open stream: %v", err)
	}

	for {
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			log.Printf("cpu error: %v", err)
			continue
		}

		metric := &pb.MetricReport{
			AgentId:     "macbook-pro",
			Timestamp:   time.Now().Unix(),
			CpuUsage:    percent[0],
			MemoryUsage: 0,
			DiskUsage:   0,
		}

		log.Printf("Sending CPU = %.2f%%", percent[0])

		if err := stream.Send(metric); err != nil {
			log.Fatalf("send error: %v", err)
		}
	}
}
