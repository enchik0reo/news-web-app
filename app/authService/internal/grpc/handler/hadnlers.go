package handler

import (
	"context"
	"errors"

	"newsWebApp/app/authService/internal/services"
	authv1 "newsWebApp/protos/gen/go/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	SaveUser(ctx context.Context, userName string, email string, password string) (int64, error)
	LoginUser(ctx context.Context, email, password string) (int64, string, string, string, error)
	Parse(ctx context.Context, acToken string) (int64, string, error)
	Refresh(ctx context.Context, refToken string) (int64, string, string, string, error)
}

type serverAPI struct {
	authv1.UnimplementedAuthServer
	authService AuthService
}

func Register(gRPC *grpc.Server, aS AuthService) {
	authv1.RegisterAuthServer(gRPC, &serverAPI{authService: aS})
}

func (s *serverAPI) SaveUser(ctx context.Context, req *authv1.SaveUserRequest) (*authv1.SaveUserResponse, error) {
	id, err := s.authService.SaveUser(ctx, req.GetUserName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidValue):
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		case errors.Is(err, services.ErrUserExists):
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.SaveUserResponse{
		UserId: id,
	}, nil
}

func (s *serverAPI) LoginUser(ctx context.Context, req *authv1.LoginUserRequest) (*authv1.LoginUserResponse, error) {
	id, userName, acsToken, refTokren, err := s.authService.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidValue):
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		case errors.Is(err, services.ErrUserDoesntExists):
			return nil, status.Error(codes.NotFound, "wrong email or password")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.LoginUserResponse{
		UserId:       id,
		UserName:     userName,
		AccessToken:  acsToken,
		RefreshToken: refTokren,
	}, nil
}

func (s *serverAPI) Parse(ctx context.Context, req *authv1.ParseRequest) (*authv1.ParseResponse, error) {
	id, userName, err := s.authService.Parse(ctx, req.GetAccessToken())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTokenExpired):
			return nil, status.Error(codes.NotFound, "token expired")
		case errors.Is(err, services.ErrInvalidToken):
			return nil, status.Error(codes.InvalidArgument, "invalid argument")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.ParseResponse{
		UserId:   id,
		UserName: userName,
	}, nil
}

func (s *serverAPI) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	id, userName, acsToken, refTokren, err := s.authService.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSessionNotFound):
			return nil, status.Error(codes.NotFound, "session expired")
		case errors.Is(err, services.ErrInvalidToken):
			return nil, status.Error(codes.InvalidArgument, "invalid argument")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.RefreshResponse{
		UserId:       id,
		UserName:     userName,
		AccessToken:  acsToken,
		RefreshToken: refTokren,
	}, nil
}
