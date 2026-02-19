package main

import (
	"gowatch/internal/database"
	"gowatch/internal/grpc"
)

func main() {
	go grpc.StartRESTServer(database.New())
	grpc.StartGRPCServer()
}
