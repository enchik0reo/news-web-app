package grpc

import (
	"context"

	authv1 "newsWebApp/protos/gen/go/auth"

	"google.golang.org/grpc"
)

type serverAPI struct {
	authv1.UnimplementedAuthServer
	authService AuthService
}

func register(gRPC *grpc.Server, aS AuthService) {
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
