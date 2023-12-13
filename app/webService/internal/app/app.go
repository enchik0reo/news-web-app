package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"newsWebApp/app/webService/internal/config"
	"newsWebApp/app/webService/internal/grpc/auth"
	"newsWebApp/app/webService/internal/http/server"
	"newsWebApp/app/webService/internal/logs"
)

type App struct {
	cfg        *config.Config
	log        *slog.Logger
	authClient *auth.Client
	srv        *server.Server
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

	handler := http.Handler

	a.srv = server.New(handler, &a.cfg.Server, a.log)

	return &a
}

func (a *App) MustRun() {
	a.log.Info("starting web service", "env", a.cfg.Env, "port", a.cfg.GRPC.Port)

	go func() {
		if err := a.srv.Start(); err != nil {
			a.log.Error("failed ower working web service", "err", err.Error())
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	a.mustStop()
}

func (a *App) mustStop() {
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.Timeout)
	defer cancel()

	if err := a.srv.Stop(ctx); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("error closing connection to web server", "err store", err.Error())
		}
	}

	a.log.Info("web service stoped gracefully")
}
