package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"newsWebApp/app/newsService/internal/grpc/handler"

	"google.golang.org/grpc"
)

type NewsService interface {
}

type Server struct {
	port       int
	log        *slog.Logger
	gRPCServer *grpc.Server
}

func New(port int, log *slog.Logger, newsService NewsService) *Server {
	grpcSrv := grpc.NewServer()

	handler.Register(grpcSrv, newsService)

	return &Server{
		port:       port,
		log:        log,
		gRPCServer: grpcSrv,
	}
}

func (s *Server) Start() error {
	const op = "grpc.server.start"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("News gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := s.gRPCServer.Serve(l); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.log.Info("Stopping news gRPC server")

	s.gRPCServer.GracefulStop()
}
