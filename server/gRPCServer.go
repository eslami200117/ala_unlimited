package server

import (
	"github.com/eslami200117/ala_unlimited/protocpb"
	"github.com/eslami200117/ala_unlimited/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func StartGRPC(core *service.Core) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	protocpb.RegisterPriceServiceServer(grpcServer, core)

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
