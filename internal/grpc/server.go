package grpc

import (
	pb "gowatch/gopherwatch/pkg/generated"
	"gowatch/internal/database"
	"io"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

var globalDB database.Service

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
}

func (s *MetricsServer) SendMetrics(stream pb.MetricsService_SendMetricsServer) error {

	for {
		metric, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&pb.Summary{
				Message: "Received metrics successfully",
			})
		}

		if err != nil {
			return err
		}

		// FAN-OUT PATTERN
		MetricChan <- metric

		log.Printf("Received from Agent: %s | CPU: %.2f",
			metric.AgentId,
			metric.CpuUsage,
		)
	}
}

type ServerInstance struct {
	GRPC *grpc.Server
}

func StartGRPCServer(db database.Service) *ServerInstance {
	StartWorkers(10, db)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(
		grpc.StreamInterceptor(ServiceIDInterceptor),
	)

	pb.RegisterMetricsServiceServer(server, &MetricsServer{})

	go func() {
		log.Println("gRPC Server running on :50051")
		if err := server.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	return &ServerInstance{
		GRPC: server,
	}
}

type RestServer struct {
	db database.Service
}

func StartRESTServer(db database.Service) {
	srv := &RestServer{db: db}
	err := http.ListenAndServe(":8080", srv.RegisterRoutes())
	if err != nil {
		log.Fatalf("REST server error: %v", err)
	}
}
