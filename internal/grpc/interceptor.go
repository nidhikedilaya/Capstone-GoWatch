package grpc

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func ServiceIDInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	md, ok := metadata.FromIncomingContext(ss.Context())
	if ok {
		ids := md.Get("service-id")
		if len(ids) > 0 {
			log.Printf("Service-ID: %s", ids[0])
		}
	}

	return handler(srv, ss)
}
