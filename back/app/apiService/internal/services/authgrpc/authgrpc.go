package authgrpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"newsWebApp/app/apiService/internal/services"
	authv1 "newsWebApp/protos/gen/go/auth"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	api authv1.AuthClient
}

func New(
	ctx context.Context,
	log *slog.Logger,
	host string,
	port int,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "authgrpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx,
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(interceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: authv1.NewAuthClient(cc),
	}, nil
}

func (c *Client) SaveUser(ctx context.Context, userName string, email string, password string) (int64, error) {
	resp, err := c.api.SaveUser(ctx, &authv1.SaveUserRequest{UserName: userName, Email: email, Password: password})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.InvalidArgument, "invalid credentials")):
			return 0, services.ErrInvalidValue
		case errors.Is(err, status.Error(codes.AlreadyExists, "user already exists")):
			return 0, services.ErrUserExists
		default:
			return 0, err
		}
	}

	return resp.UserId, nil
}

func (c *Client) LoginUser(ctx context.Context, email, password string) (int64, string, string, string, error) {
	resp, err := c.api.LoginUser(ctx, &authv1.LoginUserRequest{Email: email, Password: password})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.InvalidArgument, "invalid email or password")):
			return 0, "", "", "", services.ErrInvalidValue
		case errors.Is(err, status.Error(codes.NotFound, "wrong email or password")):
			return 0, "", "", "", services.ErrUserDoesntExists
		default:
			return 0, "", "", "", err
		}
	}

	return resp.UserId, resp.UserName, resp.AccessToken, resp.RefreshToken, nil
}

func (c *Client) Parse(ctx context.Context, acToken string) (int64, string, error) {
	resp, err := c.api.Parse(ctx, &authv1.ParseRequest{AccessToken: acToken})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.NotFound, "token expired")):
			return 0, "", services.ErrTokenExpired
		case errors.Is(err, status.Error(codes.InvalidArgument, "invalid argument")):
			return 0, "", services.ErrInvalidToken
		default:
			return 0, "", err
		}
	}

	return resp.UserId, resp.UserName, nil
}

func (c *Client) Refresh(ctx context.Context, refToken string) (int64, string, string, string, error) {
	resp, err := c.api.Refresh(ctx, &authv1.RefreshRequest{RefreshToken: refToken})
	if err != nil {
		switch {
		case errors.Is(err, status.Error(codes.NotFound, "session expired")):
			return 0, "", "", "", services.ErrSessionNotFound
		case errors.Is(err, status.Error(codes.InvalidArgument, "invalid argument")):
			return 0, "", "", "", services.ErrInvalidToken
		default:
			return 0, "", "", "", err
		}
	}

	return resp.UserId, resp.UserName, resp.AccessToken, resp.RefreshToken, nil
}

func (c *Client) CheckEmail(ctx context.Context, email string) (bool, error) {
	resp, err := c.api.CheckEmail(ctx, &authv1.CheckEmailRequest{Email: email})
	if err != nil {
		return false, err
	}

	return resp.Answer, nil
}

func (c *Client) CheckUserName(ctx context.Context, userName string) (bool, error) {
	resp, err := c.api.CheckUserName(ctx, &authv1.CheckUserNameRequest{UserName: userName})
	if err != nil {
		return false, err
	}

	return resp.Answer, nil
}

func interceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}
