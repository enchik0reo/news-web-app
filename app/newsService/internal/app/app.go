package app

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"newsWebApp/app/newsService/internal/config"
	"newsWebApp/app/newsService/internal/logs"
	"newsWebApp/app/newsService/internal/services/fetcher"
	"newsWebApp/app/newsService/internal/services/notifier"
	"newsWebApp/app/newsService/internal/storage"
	"newsWebApp/app/newsService/internal/storage/psql"
)

type App struct {
	cfg   *config.Config
	log   *slog.Logger
	db    *sql.DB
	fetch *fetcher.Fetcher
	not   *notifier.Notifier
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

	a.fetch = fetcher.New(articleStor, sourceStor, a.cfg.Manager.FetchInterval, a.cfg.Manager.FilterKeywords, a.log)

	a.not = notifier.New(articleStor, a.fetch, a.cfg.Manager.NotificationInterval, a.cfg.Manager.ArticlesLimit, a.log)

	// GRPC NEW

	return &a
}

func (a *App) MustRun() {
	a.log.Info("Starting news grpc service", "env", a.cfg.Env, "port", a.cfg.GRPC.Port)

	// GRPC START

	ctx, cacnel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cacnel()

	go func() {
		if err := a.fetch.Start(ctx); err != nil {
			a.log.Error("Working fetcher", "err store", err.Error())
		}
	}()

	go func() {
		if err := a.not.Start(ctx); err != nil {
			a.log.Error("Working notifier", "err store", err.Error())
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

	// GRPC STOP

	a.log.Info("News grpc service stoped gracefully")
}
