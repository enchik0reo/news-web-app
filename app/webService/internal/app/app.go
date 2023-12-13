package app

import (
	"context"
	"log/slog"
	"os"

	"newsWebApp/app/webService/internal/config"
	"newsWebApp/app/webService/internal/grpc/auth"
	"newsWebApp/app/webService/internal/logs"
)

type App struct {
	cfg        *config.Config
	log        *slog.Logger
	authClient *auth.Client
	//apiServer *apiserver.Server
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.Timeout)
	defer cancel()

	a.authClient, err = auth.New(ctx, a.log, a.cfg.GRPC.Port, a.cfg.GRPC.Timeout, a.cfg.GRPC.RetriesCount)
	if err != nil {
		a.log.Error("failed to create new auth client", "err", err)
		os.Exit(1)
	}

	return &a
}
