package server

import (
	"github.com/eslami200117/ala_unlimited/config"
	"github.com/eslami200117/ala_unlimited/pkg/comm"
	"github.com/eslami200117/ala_unlimited/protocpb"
	"github.com/eslami200117/ala_unlimited/service"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

type GRPCServer struct {
	logger zerolog.Logger
}

func NewGRPCServer() *GRPCServer {

	return &GRPCServer{
		logger: comm.Logger("gRPCServer"),
	}
}

func (g *GRPCServer) StartGRPC(core *service.Core, conf *config.Config) {
	lis, err := net.Listen("tcp", ":"+conf.GrpcPort)
	if err != nil {
		g.logger.Fatal().Err(err).Msg("failed to listen")
	}

	kaEnforcement := keepalive.EnforcementPolicy{
		MinTime:             60 * time.Second, // at least 60s between client pings
		PermitWithoutStream: true,             // allow ping even if no active RPCs
	}
	kaParams := keepalive.ServerParameters{
		Time:    60 * time.Second, // ping clients every 60s if idle
		Timeout: 20 * time.Second, // wait 20s for ack before closing
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaEnforcement),
		grpc.KeepaliveParams(kaParams),
	)

	protocpb.RegisterPriceServiceServer(grpcServer, core)
	g.logger.Info().Msg("gRPC server listening on :" + conf.GrpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		g.logger.Fatal().Err(err).Msg("failed to serve")
	}
}
