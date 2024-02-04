package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"newsWebApp/app/newsService/internal/config"
	grpcServer "newsWebApp/app/newsService/internal/grpc/server"
	"newsWebApp/app/newsService/internal/logs"
	"newsWebApp/app/newsService/internal/services/cacher"
	"newsWebApp/app/newsService/internal/services/fetcher"
	"newsWebApp/app/newsService/internal/services/processor"
	"newsWebApp/app/newsService/internal/storage/psql"
	"newsWebApp/app/newsService/internal/storage/redis"
	"newsWebApp/migrations/migrator"
)

type App struct {
	cfg        *config.Config
	log        *slog.Logger
	db         *sql.DB
	linkCache  *redis.Storage
	fetcher    *fetcher.Fetcher
	processor  *processor.Processor
	gRPCServer *grpcServer.Server
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	a.log.With("service", "News")

	err = migrator.Up()
	if err != nil {
		a.log.Error("Failed to apply migrations", "err", err.Error())
		os.Exit(1)
	}

	a.db, err = connectToDB(a.cfg.Storage)
	if err != nil {
		a.log.Error("Failed to create new news storage", "err", err.Error())
		os.Exit(1)
	}

	sourceStor := psql.NewSourceStorage(a.db)
	articleStor := psql.NewArticleStorage(a.db)

	ctx, cacnel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cacnel()

	a.linkCache, err = connectToLinkCache(ctx, a.cfg.LinkStorage.Host, a.cfg.LinkStorage.Port)
	if err != nil {
		a.log.Error("Failed to create new link storage", "err", err.Error())
		os.Exit(1)
	}

	linkCacher := cacher.New(a.linkCache)

	a.fetcher = fetcher.New(articleStor,
		sourceStor,
		linkCacher,
		a.cfg.Manager.FetchInterval,
		a.cfg.Manager.FilterKeywords,
		a.log,
	)

	a.processor = processor.New(articleStor,
		a.fetcher,
		a.cfg.Manager.ArticlesLimit,
		a.log,
	)

	a.gRPCServer = grpcServer.New(a.cfg.GRPC.Port, a.log, a.processor)

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
				a.log.Error("Failed ower working fetcher in news grpc service", "err", err.Error())
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
		a.log.Error("Closing connection to news storage", "err", err.Error())
	}

	if err := a.linkCache.CloseConn(); err != nil {
		a.log.Error("Closing connection to session storage", "err", err.Error())
	}

	a.gRPCServer.Stop()

	a.log.Info("News grpc service stoped gracefully")
}

func connectToDB(storage config.Postgres) (*sql.DB, error) {
	var err error
	var db *sql.DB

	for i := 1; i <= 5; i++ {
		db, err = psql.New(storage)
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("after retries: %w", err)
	}

	return db, nil
}

func connectToLinkCache(ctx context.Context, host string, port string) (*redis.Storage, error) {
	var err error
	var c *redis.Storage

	for i := 1; i <= 5; i++ {
		c, err = redis.New(ctx, host, port)
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("after retries: %w", err)
	}

	return c, nil
}
