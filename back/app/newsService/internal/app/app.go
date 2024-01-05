package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"newsWebApp/app/newsService/internal/config"
	grpcServer "newsWebApp/app/newsService/internal/grpc/server"
	"newsWebApp/app/newsService/internal/logs"
	"newsWebApp/app/newsService/internal/services/fetcher"
	"newsWebApp/app/newsService/internal/services/notifier"
	"newsWebApp/app/newsService/internal/storage"
	"newsWebApp/app/newsService/internal/storage/psql"
)

type App struct {
	cfg        *config.Config
	log        *slog.Logger
	db         *sql.DB
	fetcher    *fetcher.Fetcher
	notifier   *notifier.Notifier
	gRPCServer *grpcServer.Server
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	a.log.With("service", "News")

	a.db, err = psql.New(a.cfg.Storage)
	if err != nil {
		a.log.Error("Failed to create new news storage", "err", err)
		os.Exit(1)
	}

	sourceStor := storage.NewSourceStorage(a.db)
	articleStor := storage.NewArticleStorage(a.db)

	a.fetcher = fetcher.New(articleStor,
		sourceStor,
		a.cfg.Manager.FetchInterval,
		a.cfg.Manager.FilterKeywords,
		a.log,
	)

	a.notifier = notifier.New(articleStor,
		a.fetcher,
		a.cfg.Manager.ArticlesLimit,
		a.log,
	)

	a.gRPCServer = grpcServer.New(a.cfg.GRPC.Port, a.log, a.notifier)

	return &a
}

func (a *App) MustRun() {
	a.log.Info("Starting news grpc service", "env", a.cfg.Env, "port", a.cfg.GRPC.Port)

	go func() {
		if err := a.gRPCServer.Start(); err != nil {
			a.log.Error("Failed ower working news grpc service", "err", err.Error())
			os.Exit(1)
		}
	}()

	ctx, cacnel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cacnel()

	go func() {
		if err := a.fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				a.log.Error("Failed ower working fetcher in news grpc service", "err store", err.Error())
				os.Exit(1)
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	a.mustStop()
}

func (a *App) mustStop() {
	if err := a.db.Close(); err != nil {
		a.log.Error("Closing connection to news storage", "err store", err.Error())
	}

	a.gRPCServer.Stop()

	a.log.Info("News grpc service stoped gracefully")
}
