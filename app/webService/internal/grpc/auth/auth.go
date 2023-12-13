package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	authv1 "newsWebApp/protos/gen/go/auth"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api authv1.AuthClient
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr int,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx,
		fmt.Sprintf(":%d", addr),
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

func (c *Client) SaveUser(email string, password string) (int64, error) {
	panic("implement me")
}

func (c *Client) LoginUser(email, password string) (int64, string, string, error) {
	panic("implement me")
}

func (c *Client) Parse(acToken string) (int64, error) {
	panic("implement me")
}

func (c *Client) Refresh(refToken string) (int64, string, string, error) {
	panic("implement me")
}

func interceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}
