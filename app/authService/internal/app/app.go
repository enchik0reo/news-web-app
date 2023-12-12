package app

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"newsWebApp/app/authService/internal/config"
	grpcServer "newsWebApp/app/authService/internal/grpc/server"
	"newsWebApp/app/authService/internal/logs"
	"newsWebApp/app/authService/internal/services/auth"
	"newsWebApp/app/authService/internal/storage/psql"
	"newsWebApp/app/authService/internal/storage/redis"
)

type App struct {
	cfg         *config.Config
	log         *slog.Logger
	userStor    *psql.Storage
	sessionStor *redis.Storage
	auth        *auth.Auth
	gRPCServer  *grpcServer.Server
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	a.userStor, err = psql.New(a.cfg.UserStorage)
	if err != nil {
		a.log.Error("failed to create new user storage", "err", err)
		os.Exit(1)
	}

	a.sessionStor, err = redis.New(a.cfg.SessionStorage.Host, a.cfg.SessionStorage.Port)
	if err != nil {
		a.log.Error("failed to create new session storage", "err", err)
		os.Exit(1)
	}

	a.auth = auth.New(a.userStor, a.sessionStor, a.log, &a.cfg.Manager)

	a.gRPCServer = grpcServer.New(a.log, a.cfg.GRPC.Port, nil)

	return &a
}

func (a *App) MustRun() {
	a.log.Info("starting auth grpc service", "env", a.cfg.Env, "port", a.cfg.GRPC.Port)

	go func() {
		if err := a.gRPCServer.Start(); err != nil {
			a.log.Error("failed ower working auth grpc service", "err", err.Error())
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	a.mustStop()
}

func (a *App) mustStop() {
	if err := a.userStor.CloseConn(); err != nil {
		a.log.Error("error closing connection to user storage", "err store", err.Error())
	}

	if err := a.sessionStor.CloseConn(); err != nil {
		a.log.Error("error closing connection to session storage", "err store", err.Error())
	}

	a.gRPCServer.Stop()

	a.log.Info("auth grpc service stoped gracefully")
}
