package grpc

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type AuthService interface {
	// TODO
}

type Server struct {
	port       int
	log        *slog.Logger
	gRPCServer *grpc.Server
}

func NewServer(log *slog.Logger, port int, authService AuthService) *Server {
	grpcSrv := grpc.NewServer()

	register(grpcSrv, authService)

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

	s.log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := s.gRPCServer.Serve(l); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.log.Info("stopping gRPC server")

	s.gRPCServer.GracefulStop()
}
