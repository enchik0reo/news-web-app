package handler

import (
	"context"

	authv1 "newsWebApp/protos/gen/go/auth"

	"google.golang.org/grpc"
)

type AuthService interface {
	SaveUser(ctx context.Context, email string, password string) (int64, error)
	LoginUser(ctx context.Context, email, password string) (int64, string, string, error)
	Parse(ctx context.Context, acToken string) (int64, error)
	Refresh(ctx context.Context, refToken string) (int64, string, string, error)
}

type serverAPI struct {
	authv1.UnimplementedAuthServer
	authService AuthService
}

func Register(gRPC *grpc.Server, aS AuthService) {
	authv1.RegisterAuthServer(gRPC, &serverAPI{authService: aS})
}

func (s *serverAPI) SaveUser(ctx context.Context, req *authv1.SaveUserRequest) (*authv1.SaveUserResponse, error) {
	panic("implement me")
}

func (s *serverAPI) LoginUser(ctx context.Context, req *authv1.LoginUserRequest) (*authv1.LoginUserResponse, error) {
	panic("implement me")
}

func (s *serverAPI) Parse(ctx context.Context, req *authv1.ParseRequest) (*authv1.ParseResponse, error) {
	panic("implement me")
}

func (s *serverAPI) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	panic("implement me")
}
