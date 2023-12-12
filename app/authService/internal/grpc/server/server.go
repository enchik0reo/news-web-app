package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"newsWebApp/app/authService/internal/grpc/handler"

	"google.golang.org/grpc"
)

type Server struct {
	port       int
	log        *slog.Logger
	gRPCServer *grpc.Server
}

func New(log *slog.Logger, port int, authService handler.AuthService) *Server {
	grpcSrv := grpc.NewServer()

	handler.Register(grpcSrv, authService)

	return &Server{
		port:       port,
		log:        log,
		gRPCServer: grpcSrv,
	}
}

func (s *Server) Start() error {
	const op = "grpc.server.run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("auth gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := s.gRPCServer.Serve(l); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.log.Info("stopping auth gRPC server")

	s.gRPCServer.GracefulStop()
}
