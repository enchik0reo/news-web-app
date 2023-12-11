package app

import (
	"log/slog"

	"newsWebApp/app/authService/internal/config"
	"newsWebApp/app/authService/internal/logs"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
	/* userStor    *psql.Storage
	sessionStor *redis.Storage
	auth        *auth.Auth
	gRPCServer  *grpcsrv.Server */
}

func New() *App {
	a := App{}
	//var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	// TODO Инициализировать users хранилище

	// TODO Инициализировать sessions хранилище

	// TODO Авторизатор (Логика авторизации)

	// TODO gRPC сервер

	return &a
}

func (a *App) Run() {
	a.log.Info("starting grpc auth service", slog.Int("port", a.cfg.GRPC.Port))

	// TODO Запустить сервер приложения в отдельной горутине

	// TODO Graceful shutdown
}
