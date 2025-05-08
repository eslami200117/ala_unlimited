package server

import (
	"github.com/eslami200117/ala_unlimited/protocpb"
	"github.com/eslami200117/ala_unlimited/service"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"net"
	"os"
)

type GRPCServer struct {
	logger zerolog.Logger
}

func NewGRPCServer() *GRPCServer {
	_logger := zerolog.New(os.Stderr).
		With().Str("package", "grpc-server").
		Caller().Timestamp().Logger()

	return &GRPCServer{
		logger: _logger,
	}
}

func (g *GRPCServer) StartGRPC(core *service.Core) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		g.logger.Fatal().Err(err).Msg("failed to listen")
	}

	grpcServer := grpc.NewServer()
	protocpb.RegisterPriceServiceServer(grpcServer, core)

	g.logger.Info().Msg("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		g.logger.Fatal().Err(err).Msg("failed to serve")
	}
}
