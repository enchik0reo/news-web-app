package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"newsWebApp/app/authService/internal/grpc/handler"

	"google.golang.org/grpc"
)

type AuthService interface {
	SaveUser(ctx context.Context, userName string, email string, password string) (int64, error)
	LoginUser(ctx context.Context, email, password string) (int64, string, string, string, error)
	Parse(ctx context.Context, acToken string) (int64, string, error)
	Refresh(ctx context.Context, refToken string) (int64, string, string, string, error)
}

type RegistrService interface {
	CheckEmail(email string) (bool, error)
	CheckUserName(userName string) (bool, error)
}

type Server struct {
	port       int
	log        *slog.Logger
	gRPCServer *grpc.Server
}

func New(port int, log *slog.Logger, authService AuthService, registrService RegistrService) *Server {
	grpcSrv := grpc.NewServer()

	handler.Register(grpcSrv, authService, registrService)

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

	s.log.Info("Auth gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := s.gRPCServer.Serve(l); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.log.Info("Stopping auth gRPC server")

	s.gRPCServer.GracefulStop()
}
