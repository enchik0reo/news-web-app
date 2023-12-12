package app

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"newsWebApp/app/authService/internal/config"
	"newsWebApp/app/authService/internal/grpc"
	"newsWebApp/app/authService/internal/logs"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
	/*
		userStor    *psql.Storage
		sessionStor *redis.Storage
		auth        *auth.Auth
	*/
	gRPCServer *grpc.Server
}

func New() *App {
	a := App{}
	//var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	// TODO Инициализировать users хранилище

	// TODO Инициализировать sessions хранилище

	// TODO Auth Service

	a.gRPCServer = grpc.NewServer(a.log, a.cfg.GRPC.Port, nil)

	return &a
}

func (a *App) Run() {
	a.log.Info("starting auth grpc service", "env", a.cfg.Env, "port", a.cfg.GRPC.Port)

	go func() {
		if err := a.gRPCServer.Start(); err != nil {
			a.log.Error("failed ower working auth grpc service", "err", err.Error())
			os.Exit(1)
		}
	}()

	a.log.Info("auth grpc service is running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	a.mustStop()
}

func (a *App) mustStop() {
	/* if err := a.userStor.CloseConn(); err != nil {
		a.log.Error("error closing connection to user storage", "err store", err.Error())
	}

	if err := a.sessionStor.CloseConn(); err != nil {
		a.log.Error("error closing connection to session storage", "err store", err.Error())
	} */

	a.gRPCServer.Stop()

	a.log.Info("auth grpc service stoped gracefully")
}
